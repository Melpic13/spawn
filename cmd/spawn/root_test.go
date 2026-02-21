package main

import "testing"

func TestEvalExpression(t *testing.T) {
	t.Parallel()
	value, err := evalExpression("(2+3)*4-5/5")
	if err != nil {
		t.Fatalf("eval expression: %v", err)
	}
	if value != 19 {
		t.Fatalf("expected 19, got %v", value)
	}
}

func TestParseKV(t *testing.T) {
	t.Parallel()
	k, v, err := parseKV("region=us-east-1")
	if err != nil {
		t.Fatalf("parse kv: %v", err)
	}
	if k != "region" || v != "us-east-1" {
		t.Fatalf("unexpected key/value: %s=%s", k, v)
	}
}
