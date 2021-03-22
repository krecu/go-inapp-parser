package go_inapp_parser

import (
	"testing"
	"time"

	"github.com/k0kubun/pp"
)

func TestParseApple(t *testing.T) {
	_, err := ParseApple(
		"359917414",
		"eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlU4UlRZVjVaRFMifQ.eyJpc3MiOiI3TktaMlZQNDhaIiwiaWF0IjoxNjE2MDA5OTg1LCJleHAiOjE2MTkwMzM5ODV9.imRWFJ8QhkAfSBJFczXw4wlvsJOTuGuD8Go85hGL9gTmPCWUCktSxdtzpOrMIJGCqhvuSvXvR3Bkatfk8UR8nA",
		1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseGoogle(t *testing.T) {
	_, err := ParseGoogle("scratch.lucky.money.free.real.big.win", 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNew(t *testing.T) {
	parser := New(
		SetTimout(1*time.Second),
		SetAppleApiKey("eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IlU4UlRZVjVaRFMifQ.eyJpc3MiOiI3TktaMlZQNDhaIiwiaWF0IjoxNjE2MDA5OTg1LCJleHAiOjE2MTkwMzM5ODV9.imRWFJ8QhkAfSBJFczXw4wlvsJOTuGuD8Go85hGL9gTmPCWUCktSxdtzpOrMIJGCqhvuSvXvR3Bkatfk8UR8nA"),
	)

	app, err := parser.Parse("359917414")
	if err != nil {
		t.Fatal(err)
	}

	pp.Println(app)
}
