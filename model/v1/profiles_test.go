package model

import (
	"reflect"
	"testing"
)

func Test_splitLogs(t *testing.T) {
	type args struct {
		logText string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				"0@|0@|0@|0@|0@|0@|0@|999@2023-11-07 17:18:56,607 ERROR org.apache.juli.logging.DirectJDKLog [http-nio-19999-exec-1] Servlet.service() for servlet [dispatcherServlet] in context with path [] threw exception [Request processing failed; nested exception is org.springframework.web.client.HttpServerErrorException: 500 null] with root cause\norg.springframework.web.client.HttpServerErrorException: 500 null\n\tat org.springframework.web.client.DefaultResponseErrorHandler.handleError(DefaultResponseErrorHandler.java:97)\n\tat org.springframework.web.client.DefaultResponseErrorHandler.handleError(DefaultResponseErrorHandler.java:79)\n\tat org.springframework.web.client.ResponseErrorHandler.handleError(ResponseErrorHandler.java:63)\n\tat org.springframework.web.client.RestTemplate.handleResponse$original$19HdDxif(RestTemplate.java:775)\n\tat org.springframework.web.client.RestTemplate.handleResponse$original$19HdDxif$accessor$SNBRA5qi(RestTemplate.java)\n\tat org.springframework.web.client.RestTemplate$auxiliary$woelv6Ky.call(Unkno|",
			},
			want: []string{
				"2023-11-07 17:18:56,607 ERROR org.apache.juli.logging.DirectJDKLog [http-nio-19999-exec-1] Servlet.service() for servlet [dispatcherServlet] in context with path [] threw exception [Request processing failed; nested exception is org.springframework.web.client.HttpServerErrorException: 500 null] with root cause\norg.springframework.web.client.HttpServerErrorException: 500 null\n\tat org.springframework.web.client.DefaultResponseErrorHandler.handleError(DefaultResponseErrorHandler.java:97)\n\tat org.springframework.web.client.DefaultResponseErrorHandler.handleError(DefaultResponseErrorHandler.java:79)\n\tat org.springframework.web.client.ResponseErrorHandler.handleError(ResponseErrorHandler.java:63)\n\tat org.springframework.web.client.RestTemplate.handleResponse$original$19HdDxif(RestTemplate.java:775)\n\tat org.springframework.web.client.RestTemplate.handleResponse$original$19HdDxif$accessor$SNBRA5qi(RestTemplate.java)\n\tat org.springframework.web.client.RestTemplate$auxiliary$woelv6Ky.call(Unkno",
			},
		},
		{
			name: "test1",
			args: args{
				"0@|6@2012-1|",
			},
			want: []string{
				"2012-1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitLogs(tt.args.logText); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitLogs() = %v, want %v", got, tt.want)
			}
		})
	}
}
