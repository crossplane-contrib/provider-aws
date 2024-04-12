/*
Copyright 2024 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	smithytime "github.com/aws/smithy-go/time"
	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/crossplane-contrib/provider-aws/pkg/monitor"
)

type Monitor struct {
	client         *sqs.Client
	receiveOptions sqs.ReceiveMessageInput
	logger         logging.Logger
	subscribers    []monitor.Subscriber
	prepareHooks   []monitor.PrepareHook
}

var _ monitor.Monitor = &Monitor{}

func NewMonitor(
	config aws.Config,
	receiveOptions sqs.ReceiveMessageInput,
	logger logging.Logger,
) *Monitor {
	return &Monitor{
		client:         sqs.NewFromConfig(config),
		receiveOptions: receiveOptions,
		logger:         logger.WithValues("monitor", "SQSMonitor"),
	}
}

// AddSubscriber implements monitor.Monitor.
func (monitor *Monitor) AddSubscriber(subscriber monitor.Subscriber) {
	monitor.subscribers = append(monitor.subscribers, subscriber)
}

// AddPrepareHook implements monitor.Monitor.
func (monitor *Monitor) AddPrepareHook(hook monitor.PrepareHook) {
	monitor.prepareHooks = append(monitor.prepareHooks, hook)
}

// Prepare implements monitor.Monitor.
func (monitor *Monitor) Prepare(ctx context.Context) error {
	for i, hook := range monitor.prepareHooks {
		if err := hook(ctx); err != nil {
			return fmt.Errorf("running hook %d: %w", i, err)
		}
	}
	return nil
}

// Start implements monitor.Monitor.
func (monitor *Monitor) Start(ctx context.Context) error {
	for ctx.Err() == nil {
		// SQS supports long polling and timeout is configured with receiveOptions.WaitTimeSeconds
		response, err := monitor.client.ReceiveMessage(ctx, &monitor.receiveOptions)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		} else if err != nil {
			monitor.logger.Info("failed to receive message, pausing for 5s", "error", err)
			_ = smithytime.SleepWithContext(ctx, 5*time.Second)
			continue
		}
		// monitor.logger.Debug("received messages", "count", len(response.Messages))
		for _, message := range response.Messages {
			if err := monitor.publish(ctx, message); err != nil {
				monitor.logger.Info("failed to publish message", "error", err)
			}
			if _, err := monitor.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
				QueueUrl:      monitor.receiveOptions.QueueUrl,
				ReceiptHandle: message.ReceiptHandle,
			}); err != nil {
				monitor.logger.Info("failed to delete message", "error", err)
			}
		}
	}
	return nil
}

func (monitor *Monitor) publish(ctx context.Context, message types.Message) error {
	// monitor.logger.Debug("parsing events", "body", message.Body)
	events, err := parseEventsFromMessage(message)
	if err != nil {
		return err
	}
	for _, event := range events {
		// monitor.logger.Debug("publishing event", "event", event)
		for _, subscriber := range monitor.subscribers {
			subscriber(ctx, event)
		}
	}
	return nil
}
