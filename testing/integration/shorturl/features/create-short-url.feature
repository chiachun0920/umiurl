Feature: User create short url

    Scenario: User create short url successfully
        When I create a short url with "https://www.example.com"
        Then I should receive a short url