Feature: Round Trip Time
  In order to be confident about a responsiveness
  As an operator
  I want to be sure the server does not introduce a noticeable delay

  Scenario: Single User Performance
    Given A fake ArduMower is connected to the relay server
    When A fake Sunray app sends consecutive commands for 10 seconds
    Then The average RTT is less than 300 milliseconds
    And The error rate is below 0.001 %

  Scenario: Multi User Performance
    Given 100 fake ArduMowers are connected to the relay server
    When 100 fake Sunray apps send consecutive commands for 10 seconds
    Then The average RTT is less than 350 milliseconds
    And The error rate is below 0.01 %
