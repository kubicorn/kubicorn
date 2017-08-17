// Copyright © 2017 The Kubicorn Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package retry

import (
	"fmt"
	"testing"
)

type testStructHappy struct{}

func (t testStructHappy) Try() error {
	return nil
}

func TestRetryHappy(t *testing.T) {
	tsh := testStructHappy{}

	r := NewRetrier(10, 5, tsh)
	err := r.RunRetry()

	if err != nil {
		t.Errorf("unexpected error occurred: %#v", err)
	}
}

type testStructSad struct{}

func (t testStructSad) Try() error {
	return fmt.Errorf("error")
}

func TestRetrySad(t *testing.T) {
	tss := testStructSad{}

	r := NewRetrier(3, 1, tss)
	err := r.RunRetry()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	want := fmt.Errorf("unable to succeed at retry after 3 attempts at 1 seconds")
	if err.Error() != want.Error() {
		t.Errorf("unexpected error\n\tgot: %#v\n\twant: %#v", err, want)
	}
}
