package launcher

import (
	"reflect"
	"testing"
)

func Test_getLogLevel(t *testing.T) {
	tests := []struct {
		name string
		line string
		want LogLevel
	}{
		{
			"info",
			"2021-10-23 12:31:19 <ZWELS:50791547> TS3105 INFO (start-component.sh:95) starting component explorer-mvs ...",
			LogLevelInfo,
		},
		{
			"error",
			"2021-10-23 12:31:19 <ZWELS:50791547> TS3105 ERROR (start-component.sh:95) starting component explorer-mvs ...",
			LogLevelError,
		},
		{
			"warning",
			"2021-10-23 12:31:19 <ZWELS:50791547> TS3105 WARN (start-component.sh:95) starting component explorer-mvs ...",
			LogLevelWarning,
		},
		{
			"debug",
			"2021-10-23 12:31:19 <ZWELS:50791547> TS3105 DEBUG (start-component.sh:95) starting component explorer-mvs ...",
			LogLevelDebug,
		},
		{
			"any",
			"hello world",
			LogLevelAny,
		},
		{
			"warn with ms",
			"2021-10-24 14:30:03.059 <ZWEAAC1:InstanceLookupExecutor-Thread:459380> TS3105 WARN  (o.z.a.a.s.c.CachedProductFamilyService) ZWEAC705W The home page url for service IBMZOSMF was not transformed. Not able to select route for url of the service ibmzosmf. Original url used.",
			LogLevelWarning,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLogLevel(tt.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLogLevel() = %s, want %s", got, tt.want)
			}
		})
	}
}
