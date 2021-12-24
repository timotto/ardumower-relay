package main

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	. "github.com/timotto/ardumower-relay/test/fake_app"
	. "github.com/timotto/ardumower-relay/test/fake_ardumower"
	. "github.com/timotto/ardumower-relay/test/testbed"
	"io/ioutil"
	"net/http"
)

type feature struct {
	bed   *Testbed
	mower *FakeArdumower
	app   *FakeApp

	lastResponse *http.Response
}

func (f *feature) initialize(ctx *godog.ScenarioContext) {
	f.bed = FromEnv()
	f.mower = NewFakeArdumower(f.bed)
	f.app = NewFakeApp(f.bed)

	ctx.Before(f.before)
	ctx.After(f.after)

	ctx.Step(`^A fake ArduMower is connected to the relay server$`, f.aFakeArduMowerIsConnectedToTheRelayServer)
	ctx.Step(`^A fake Sunray app sends a command$`, f.aFakeSunrayAppSendsACommand)
	ctx.Step(`^The fake Sunray app receives the expected response$`, f.theFakeSunrayAppReceivesTheExpectedResponse)
}

func (f *feature) before(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
	return ctx, nil
}

func (f *feature) after(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
	f.mower.Stop()

	return ctx, nil
}

func (f *feature) aFakeArduMowerIsConnectedToTheRelayServer() error {
	return f.mower.Start()
}

const (
	specificCommand  = "AT+HELLO\n"
	specificResponse = "OK=HELLO\n"
)

func (f *feature) aFakeSunrayAppSendsACommand() error {
	response, err := f.app.Send(specificCommand)
	f.lastResponse = response

	return err
}

func (f *feature) theFakeSunrayAppReceivesTheExpectedResponse() error {
	if f.lastResponse == nil {
		return fmt.Errorf("there is no response")
	}

	if f.lastResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("the response has an unexpected response status: %v", f.lastResponse.Status)
	}

	body, err := ioutil.ReadAll(f.lastResponse.Body)
	if err != nil {
		return fmt.Errorf("unable to read response body: %w", err)
	}

	actualResponse := string(body)
	expectedResponse := specificResponse
	if actualResponse != expectedResponse {
		return fmt.Errorf("the actual response is different from the expected response: %v != %v", actualResponse, expectedResponse)
	}

	return nil
}
