package main

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	. "github.com/timotto/ardumower-relay/test/fake_app"
	. "github.com/timotto/ardumower-relay/test/fake_ardumower"
	. "github.com/timotto/ardumower-relay/test/testbed"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type feature struct {
	bed    *Testbed
	users  users
	mowers []*FakeArdumower
	apps   []*FakeApp
	state  state
}

type state struct {
	lastResponse *http.Response
	result       *rttMeasurementRun
}

func (f *feature) initialize(ctx *godog.ScenarioContext) {
	f.bed = FromEnv()

	var err error
	if f.users, f.mowers, f.apps, err = loadUsers(f.bed); err != nil {
		panic(fmt.Errorf(`you may need to run "go generate ./...": %w`, err))
	}

	ctx.Before(f.before)
	ctx.After(f.after)

	ctx.Step(`^A fake ArduMower is connected to the relay server$`, f.aFakeArduMowerIsConnectedToTheRelayServer)
	ctx.Step(`^A fake Sunray app sends consecutive commands for (\d+) seconds$`, f.aFakeSunrayAppSendsConsecutiveCommandsForSeconds)
	ctx.Step(`^The average RTT is less than (\d+) milliseconds$`, f.theAverageRTTIsMilliseconds)
	ctx.Step(`^The error rate is below (.+) %$`, f.theErrorRateIsBelow)

	ctx.Step(`^(\d+) fake ArduMowers are connected to the relay server$`, f.fakeArduMowersAreConnectedToTheRelayServer)
	ctx.Step(`^(\d+) fake Sunray apps send consecutive commands for (\d+) seconds$`, f.fakeSunrayAppsSendConsecutiveCommandsForSeconds)

}

func (f *feature) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	f.state = state{}

	return ctx, nil
}

func (f *feature) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	for _, mower := range f.mowers {
		mower.Stop()
	}

	return ctx, nil
}

func (f *feature) aFakeArduMowerIsConnectedToTheRelayServer() error {
	if err := f.mowers[0].Start(); err != nil {
		return err
	}

	return nil
}

func (f *feature) aFakeSunrayAppSendsConsecutiveCommandsForSeconds(maxDurationInSeconds int) error {
	maxTestDuration := time.Second * time.Duration(maxDurationInSeconds)

	f.state.result = runRttMeasurement(f.apps[0], maxTestDuration, 0)

	return nil
}

func (f *feature) theAverageRTTIsMilliseconds(arg1 int) error {
	if f.state.result == nil {
		return fmt.Errorf("there is no result")
	}

	maxAvg := time.Millisecond * time.Duration(arg1)

	actualAvg := f.state.result.AvgRtt
	if actualAvg > maxAvg {
		return fmt.Errorf("the actual average round trip time time is %v", actualAvg.String())
	}

	log.Printf("rtt avg / min / max / err-rate: %v / %v / %v / %.3f %%",
		f.state.result.AvgRtt, f.state.result.MinRtt, f.state.result.MaxRtt, f.state.result.ErrorRate*100)

	return nil
}

func (f *feature) theErrorRateIsBelow(val string) error {
	acceptableErrorRate, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}

	actualErrorRate := f.state.result.ErrorRate * 100
	if actualErrorRate > acceptableErrorRate {
		return fmt.Errorf("the actual error rate of %.3f %% is higher than the acceptable error rate of %.4f %%", acceptableErrorRate, acceptableErrorRate)
	}

	return nil
}

func (f *feature) fakeArduMowersAreConnectedToTheRelayServer(count int) error {
	if count > len(f.users) {
		return fmt.Errorf(`there are only %v credentials available - increase "%v" and run "go generate ./..." for %v users`, len(f.users), "RELAY_LOADTEST_USER_COUNT", count)
	}

	for i := 0; i < count; i++ {
		if err := f.mowers[i].Start(); err != nil {
			return fmt.Errorf("failed to start fake ArduMower # %v: %w", i, err)
		}
	}

	return nil
}

func (f *feature) fakeSunrayAppsSendConsecutiveCommandsForSeconds(count, maxDurationInSeconds int) error {
	if count > len(f.mowers) {
		return fmt.Errorf("there are %v mowers but you want %v apps", f.mowers, count)
	}

	maxTestDuration := time.Second * time.Duration(maxDurationInSeconds)

	wg := &sync.WaitGroup{}
	wg.Add(count)
	results := make([]*rttMeasurementRun, count)

	for i := 0; i < count; i++ {
		go func(index int) {
			defer wg.Done()
			results[index] = runRttMeasurement(f.apps[index], maxTestDuration, 0)
		}(i)
	}

	wg.Wait()
	result := &rttMeasurementRun{}
	result.Add(results...)

	f.state.result = result

	return nil
}
