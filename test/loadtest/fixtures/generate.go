package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

//go:generate go run .

const (
	defaultCount = 1000
	countEnvKey  = "RELAY_LOADTEST_USER_COUNT"

	filename = "users.txt"

	defaultUsernameFmt = "loadtest-user-%v"
	usernameFmtEnvKey  = "RELAY_LOADTEST_USERNAME_FMT"

	defaultPasswordFmt = "default-loadtest-password-%v"
	passwordFmtEnvKey  = "RELAY_LOADTEST_PASSWORD_FMT"
)

var (
	count       = lookupDesiredEntryCount()
	usernameGen = lookupUsernameGenerator()
	passwordGen = lookupPasswordGenerator()
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(file)

	log.Printf("writing %v", filename)
	if err := generator(w.WriteString); err != nil {
		return err
	}

	if err := w.Flush(); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func generator(writer func(string) (int, error)) error {
	for i := 0; i < count; i++ {
		if _, err := writer(generateLine(i)); err != nil {
			return fmt.Errorf("failed to write line %v: %w", i+1, err)
		}
	}

	return nil
}

func generateLine(index int) string {
	username := usernameGen(index)
	password := passwordGen(index)

	return fmt.Sprintf("%v:%v\n", username, password)
}

func lookupDesiredEntryCount() int {
	if val, found := os.LookupEnv(countEnvKey); !found {
		return defaultCount
	} else if i, err := strconv.ParseInt(val, 10, 32); err != nil {
		panic(fmt.Errorf("%v is not a valid integer: %w", countEnvKey, err))
	} else {
		return int(i)
	}
}

type intFormatter func(int) string

func lookupUsernameGenerator() intFormatter {
	format := lookupEnvOrWorry(usernameFmtEnvKey, defaultUsernameFmt)

	if hasValuePlaceholder(format) {
		return numberedStringGenerator(format)
	}

	panic(usernameFmtEnvKey + " must contain a %v placeholder")
}

func lookupPasswordGenerator() intFormatter {
	passwordFmt := lookupEnvOrWorry(passwordFmtEnvKey, defaultPasswordFmt)

	if hasValuePlaceholder(passwordFmt) {
		return numberedStringGenerator(passwordFmt)
	}

	log.Println("all generated passwords will be the same")

	return func(_ int) string {
		return passwordFmt
	}
}

func lookupEnvOrWorry(key, fallback string) string {
	if val, found := os.LookupEnv(key); found {
		return val
	}

	log.Printf("change %v if you want to use the generated credentials on the Internet", key)

	return fallback

}

func hasValuePlaceholder(format string) bool {
	return strings.HasSuffix(fmt.Sprintf(format), "(MISSING)")
}

func numberedStringGenerator(passwordFmt string) intFormatter {
	return func(index int) string {
		return fmt.Sprintf(passwordFmt, index)
	}
}
