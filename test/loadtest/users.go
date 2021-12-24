package main

import (
	"fmt"
	. "github.com/timotto/ardumower-relay/test/fake_app"
	. "github.com/timotto/ardumower-relay/test/fake_ardumower"
	"github.com/timotto/ardumower-relay/test/testbed"
	"os"
	"strings"
)

type users []*user

type user struct {
	Username string
	Password string
}

func loadUsers(bed *testbed.Testbed) ([]*user, []*FakeArdumower, []*FakeApp, error) {
	filename := "test/loadtest/fixtures/users.txt"

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot read %v: %w", filename, err)
	}

	var result []*user
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			return nil, nil, nil, fmt.Errorf("invalid entry in line %v: %v", i+1, line)
		}

		usr := &user{
			Username: parts[0],
			Password: strings.Join(parts[1:], ":"),
		}

		result = append(result, usr)
	}

	count := len(result)
	if count == 0 {
		return nil, nil, nil, fmt.Errorf("the user.txt file cannot be empty")
	}

	var mowers []*FakeArdumower
	var apps []*FakeApp
	for i := 0; i < count; i++ {
		u := result[i]
		b := &testbed.Testbed{
			RelayServerUrl: bed.RelayServerUrl,
			Username:       u.Username,
			Password:       u.Password,
		}

		mower := NewFakeArdumower(b)
		app := NewFakeApp(b)

		mowers = append(mowers, mower)
		apps = append(apps, app)
	}

	return result, mowers, apps, nil
}
