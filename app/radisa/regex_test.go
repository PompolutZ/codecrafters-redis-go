package radisa

import (
	"reflect"
	"testing"
)

func TestMatchAll(t *testing.T) {
	result := SearchKeys("*", []string{"foo", "bar"})
	expected := []string{"foo", "bar"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestNoMatch(t *testing.T) {
	result := SearchKeys("baz", []string{"foo", "bar"})
	expected := []string{}

	if len(result) > 0 {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestExactMatch(t *testing.T) {
	result := SearchKeys("foo", []string{"foo", "bar"})
	expected := []string{"foo"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPartialMatchWithWild(t *testing.T) {
	result := SearchKeys("ba*", []string{"foo", "bar", "baz"})
	expected := []string{"bar", "baz"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestGlobInTheMiddle(t *testing.T) {
	result := SearchKeys("f*a", []string{"faa", "fooa", "fxyza", "bar", "baz"})
	expected := []string{"faa", "fooa", "fxyza"}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}	