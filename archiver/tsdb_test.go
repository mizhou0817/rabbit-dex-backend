package archiver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"
)

func TestMultiErrInDefer(t *testing.T) {
	check := func(commitErr, rollbackErr error) (err error) {
		defer func() {
			if err != nil {
				err = multierr.Append(err, rollbackErr)
			}
		}()
		err = commitErr
		return err
	}

	commitError := errors.New("commit error")
	rollbackError := errors.New("rollback error")

	cases := []struct {
		desc     string
		commit   error
		rollback error
		wantErr  error
	}{
		{
			desc:     "nil",
			commit:   nil,
			rollback: nil,
			wantErr:  nil,
		},
		{
			desc:     "commit errored, rollback not => commit error",
			commit:   commitError,
			rollback: nil,
			wantErr:  commitError,
		},
		{
			desc:     "commit succeeded, rollback errored => no error",
			commit:   nil,
			rollback: commitError,
			wantErr:  nil,
		},
		{
			desc:     "commit and rollback errored => mutierr",
			commit:   commitError,
			rollback: rollbackError,
			wantErr:  multierr.Append(commitError, rollbackError),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := check(tc.commit, tc.rollback)
			if tc.wantErr != nil && !assert.Equal(t, err, tc.wantErr) {
				t.Errorf("got %v, want %v", err, tc.wantErr)
			}
		})
	}
}
