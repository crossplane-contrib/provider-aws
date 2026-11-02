/*
Copyright 2026 The Crossplane Authors.

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

package broker

import (
	"context"
	"testing"

	svcsdk "github.com/aws/aws-sdk-go/service/mq"
	"github.com/google/go-cmp/cmp"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mq/v1alpha1"
)

func TestIsUpToDate(t *testing.T) {
	engineVersion := "3.10.25"
	newEngineVersion := "3.11.16"

	hostInstanceType := "mq.t3.micro"
	newHostInstanceType := "mq.m5.large"

	dayOfWeek := "MONDAY"
	newDayOfWeek := "TUESDAY"
	timeOfDay := "03:00"
	timeZone := "UTC"

	authStrategy := "SIMPLE"
	newAuthStrategy := "LDAP"
	yes, no := true, false

	type args struct {
		broker                 *svcapitypes.Broker
		describeBrokerResponse *svcsdk.DescribeBrokerResponse
	}
	type want struct {
		result bool
	}
	cases := map[string]struct {
		args
		want
	}{
		"NothingChanged": {
			args: args{
				broker: &svcapitypes.Broker{
					Spec: svcapitypes.BrokerSpec{
						ForProvider: svcapitypes.BrokerParameters{
							EngineVersion:    &engineVersion,
							HostInstanceType: &hostInstanceType,
							MaintenanceWindowStartTime: &svcapitypes.WeeklyStartTime{
								DayOfWeek: &dayOfWeek,
								TimeOfDay: &timeOfDay,
								TimeZone:  &timeZone,
							},
						},
					},
				},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{
					EngineVersion:    &engineVersion,
					HostInstanceType: &hostInstanceType,
					MaintenanceWindowStartTime: &svcsdk.WeeklyStartTime{
						DayOfWeek: &dayOfWeek,
						TimeOfDay: &timeOfDay,
						TimeZone:  &timeZone,
					},
				},
			},
			want: want{result: true},
		},
		"EngineVersionChanged": {
			args: args{
				broker: &svcapitypes.Broker{
					Spec: svcapitypes.BrokerSpec{
						ForProvider: svcapitypes.BrokerParameters{
							EngineVersion:    &newEngineVersion,
							HostInstanceType: &hostInstanceType,
						},
					},
				},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{
					EngineVersion:    &engineVersion,
					HostInstanceType: &hostInstanceType,
				},
			},
			want: want{result: false},
		},
		"HostInstanceTypeChanged": {
			args: args{
				broker: &svcapitypes.Broker{
					Spec: svcapitypes.BrokerSpec{
						ForProvider: svcapitypes.BrokerParameters{
							EngineVersion:    &engineVersion,
							HostInstanceType: &newHostInstanceType,
						},
					},
				},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{
					EngineVersion:    &engineVersion,
					HostInstanceType: &hostInstanceType,
				},
			},
			want: want{result: false},
		},
		"MaintenanceWindowChanged": {
			args: args{
				broker: &svcapitypes.Broker{
					Spec: svcapitypes.BrokerSpec{
						ForProvider: svcapitypes.BrokerParameters{
							EngineVersion: &engineVersion,
							MaintenanceWindowStartTime: &svcapitypes.WeeklyStartTime{
								DayOfWeek: &newDayOfWeek,
								TimeOfDay: &timeOfDay,
								TimeZone:  &timeZone,
							},
						},
					},
				},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{
					EngineVersion: &engineVersion,
					MaintenanceWindowStartTime: &svcsdk.WeeklyStartTime{
						DayOfWeek: &dayOfWeek,
						TimeOfDay: &timeOfDay,
						TimeZone:  &timeZone,
					},
				},
			},
			want: want{result: false},
		},
		"PartiallySpecifiedMaintenanceWindowMatchesObserved": {
			// A caller may only set DayOfWeek, leaving TimeOfDay/TimeZone
			// unset. Since GenerateUpdateBrokerRequest never sends unset
			// sub-fields to AWS, isUpToDate must not flag a diff for them
			// either or the resource would be stuck in a permanent,
			// unfixable "not up to date" loop.
			args: args{
				broker: &svcapitypes.Broker{
					Spec: svcapitypes.BrokerSpec{
						ForProvider: svcapitypes.BrokerParameters{
							EngineVersion: &engineVersion,
							MaintenanceWindowStartTime: &svcapitypes.WeeklyStartTime{
								DayOfWeek: &dayOfWeek,
							},
						},
					},
				},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{
					EngineVersion: &engineVersion,
					MaintenanceWindowStartTime: &svcsdk.WeeklyStartTime{
						DayOfWeek: &dayOfWeek,
						TimeOfDay: &timeOfDay,
						TimeZone:  &timeZone,
					},
				},
			},
			want: want{result: true},
		},
		"PendingEngineVersionMatchesSpec": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{EngineVersion: &newEngineVersion}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{EngineVersion: &engineVersion, PendingEngineVersion: &newEngineVersion},
			},
			want: want{result: true},
		},
		"PendingEngineVersionDiffersFromSpec": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{EngineVersion: &engineVersion}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{EngineVersion: &engineVersion, PendingEngineVersion: &newEngineVersion},
			},
			want: want{result: false},
		},
		"PendingHostInstanceTypeMatchesSpec": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{HostInstanceType: &newHostInstanceType}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{HostInstanceType: &hostInstanceType, PendingHostInstanceType: &newHostInstanceType},
			},
			want: want{result: true},
		},
		"PendingAuthenticationStrategyMatchesSpec": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{AuthenticationStrategy: &newAuthStrategy}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{AuthenticationStrategy: &authStrategy, PendingAuthenticationStrategy: &newAuthStrategy},
			},
			want: want{result: true},
		},
		"PendingAuthenticationStrategyDiffersFromSpec": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{AuthenticationStrategy: &authStrategy}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{AuthenticationStrategy: &authStrategy, PendingAuthenticationStrategy: &newAuthStrategy},
			},
			want: want{result: false},
		},
		"PendingLogsMatchSpec": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{Logs: &svcapitypes.Logs{Audit: &yes}}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{Logs: &svcsdk.LogsSummary{Audit: &no, Pending: &svcsdk.PendingLogs{Audit: &yes}}},
			},
			want: want{result: true},
		},
		"PendingLogsDifferFromSpec": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{Logs: &svcapitypes.Logs{Audit: &no}}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{Logs: &svcsdk.LogsSummary{Audit: &no, Pending: &svcsdk.PendingLogs{Audit: &yes}}},
			},
			want: want{result: false},
		},
		"AutoMinorVersionUpgradeSkipsEngineVersion": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{AutoMinorVersionUpgrade: &yes, EngineVersion: &engineVersion}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{AutoMinorVersionUpgrade: &yes, EngineVersion: &newEngineVersion},
			},
			want: want{result: true},
		},
		"AutoMinorVersionUpgradeDisabledComparesEngineVersion": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{AutoMinorVersionUpgrade: &no, EngineVersion: &engineVersion}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{AutoMinorVersionUpgrade: &no, EngineVersion: &newEngineVersion},
			},
			want: want{result: false},
		},
		"AutoMinorVersionUpgradeChanged": {
			args: args{
				broker:                 &svcapitypes.Broker{Spec: svcapitypes.BrokerSpec{ForProvider: svcapitypes.BrokerParameters{AutoMinorVersionUpgrade: &yes}}},
				describeBrokerResponse: &svcsdk.DescribeBrokerResponse{AutoMinorVersionUpgrade: &no},
			},
			want: want{result: false},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result, _, err := isUpToDate(context.TODO(), tc.args.broker, tc.args.describeBrokerResponse)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want.result, result); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
