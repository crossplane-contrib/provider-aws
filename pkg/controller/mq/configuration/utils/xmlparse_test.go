package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDiffMQConfiguration(t *testing.T) {
	type args struct {
		in  string
		out string
	}

	type want struct {
		diff string
	}

	cases := map[string]struct {
		args
		want
	}{
		"IdenticalConfigs with unsorted attributes": {
			args: args{
				in: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
                     <broker xmlns="http://activemq.apache.org/schema/core">
                       <plugins>
                         <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
                         <statisticsBrokerPlugin/>
                        <timeStampingBrokerPlugin zeroExpirationOverride="86400000" ttlCeiling="84400000"/>
                       </plugins>
                     </broker>`,
				out: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
                      <broker xmlns="http://activemq.apache.org/schema/core">
                        <plugins>
                          <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
                          <statisticsBrokerPlugin/>
                          <timeStampingBrokerPlugin ttlCeiling="84400000" zeroExpirationOverride="86400000"/>
                        </plugins>
                      </broker>`,
			},
			want: want{
				diff: "",
			},
		},
		"IdenticalConfigs, ignore comments": {
			args: args{
				in: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		                 <broker xmlns="http://activemq.apache.org/schema/core">
              <!--
                <destinationInterceptors>
                  <mirroredQueue copyMessage="true" postfix=".qmirror" prefix=""/>
                  <virtualDestinationInterceptor>
                    <virtualDestinations>
                      <virtualTopic name="&gt;" prefix="VirtualTopicConsumers.*." selectorAware="false"/>
                      <compositeQueue name="MY.QUEUE">
                        <forwardTo>
                          <queue physicalName="FOO"/>
                          <topic physicalName="BAR"/>
                        </forwardTo>
                      </compositeQueue>
                    </virtualDestinations>
                  </virtualDestinationInterceptor>
                </destinationInterceptors>
               -->
		                    <plugins>
		                      <forcePersistencyModeBrokerPlugin persistenceFlag="false"/>
		                      <statisticsBrokerPlugin/>
                          <timeStampingBrokerPlugin zeroExpirationOverride="86400001" ttlCeiling="84400001"/>
		                    </plugins>
		                 </broker>`,
				out: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
		                  <broker xmlns="http://activemq.apache.org/schema/core">
		                    <plugins>
		                      <forcePersistencyModeBrokerPlugin persistenceFlag="false"/>
		                      <statisticsBrokerPlugin/>
                          <timeStampingBrokerPlugin zeroExpirationOverride="86400001" ttlCeiling="84400001"/>
		                    </plugins>
		                  </broker>`,
			},
			want: want{
				diff: "",
			},
		},
		"Same config with diff tag order and wrong attribute order": {
			args: args{
				in: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
                     <broker xmlns="http://activemq.apache.org/schema/core">
                       <plugins>
                        <statisticsBrokerPlugin/>
                         <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
                        <timeStampingBrokerPlugin zeroExpirationOverride="86400000" ttlCeiling="84400000"/>
                       </plugins>
                     </broker>`,
				out: `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
                      <broker xmlns="http://activemq.apache.org/schema/core">
                        <plugins>
                          <forcePersistencyModeBrokerPlugin persistenceFlag="true"/>
                          <statisticsBrokerPlugin/>
                          <timeStampingBrokerPlugin ttlCeiling="84400000" zeroExpirationOverride="86400000"/>
                        </plugins>
                      </broker>`,
			},
			want: want{
				diff: "",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			diff, err := DiffXMLConfigs(tc.args.in, tc.args.out)
			if diff := cmp.Diff(tc.want.diff, diff); diff != "" || err != nil {
				t.Errorf("DiffXMLConfigs() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
