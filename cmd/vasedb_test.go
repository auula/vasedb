package main

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	testCases := []struct {
		input    []string
		expected []string
	}{
		{
			input:    []string{"./vasedb", "--port=2468", "--host=localhost", "--flag", "value"},
			expected: []string{"--port", "2468", "--host", "localhost", "--flag", "value"},
		},
		{
			input:    []string{"./vasedb", "--port==8080", "--port===8080", "--flag=value"},
			expected: []string{"--flag", "value"},
		},
		{
			input:    []string{"./vasedb", "arg1", "arg2", "arg3"},
			expected: []string{"arg1", "arg2", "arg3"},
		},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			result := splitArgs(testCase.input)
			if !reflect.DeepEqual(result, testCase.expected) {
				t.Errorf("Expected %v, but got %v", testCase.expected, result)
			}
		})
	}
}

func TestTrimDaemon(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		// 测试移除 "-daemon" 参数的情况
		{
			input:    []string{"app", "-daemon", "arg1", "arg2", "--daemon", "arg3"},
			expected: []string{"arg1", "arg2", "arg3"},
		},
		// 测试不包含 "-daemon" 参数的情况
		{
			input:    []string{"app", "arg1", "arg2", "arg3"},
			expected: []string{"arg1", "arg2", "arg3"},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			result := trimDaemon(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}
