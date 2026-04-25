//go:build e2e

package tests

import "testing"

func TestRequestEmptyBody(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_server_returning_empty_body()

	when.
		a_jsonnet_request_is_evaluated("GET", "/empty")

	then.the_eval_has_no_error().and().
		the_result_is_an_empty_map()
}

func TestRequestHTTPError(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_server_returning_not_found("gone")

	when.
		a_jsonnet_request_is_evaluated("GET", "/missing")

	then.
		the_eval_has_no_error().and().
		the_result_has_kind("Status").and().
		the_result_has_code(404).and().
		the_result_has_message("gone")
}

func TestRequestTimeout(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_slow_server_and_short_timeout()

	when.
		a_jsonnet_request_is_evaluated("GET", "/slow")

	then.
		the_eval_has_no_error().and().
		the_result_has_kind("Status").and().
		the_result_has_code(500).and().
		the_result_message_contains("context deadline exceeded")
}

func TestRequestPassesQueryAndHeaders(t *testing.T) {
	given, when, then := scenario(t)

	given.
		a_server_echoing_query_and_header()

	when.
		a_jsonnet_request_with_query_and_headers_is_evaluated("GET", "/echo", "up", "hello")

	then.
		the_eval_has_no_error().and().
		the_result_has_field("query", "up").and().
		the_result_has_field("header", "hello")
}
