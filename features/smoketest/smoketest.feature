Feature: Smoke Test
  In order to be confident about a deployment
  As an operator
  I want to be sure the happy path works as expected

  Scenario: Echo server responds
    Given A fake ArduMower is connected to the relay server
    When A fake Sunray app sends a command
    Then The fake Sunray app receives the expected response
