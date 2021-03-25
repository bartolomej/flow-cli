/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package cli defines constants, configurations, and utilities that are used across the Flow CLI.
package util

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/onflow/flow-go-sdk/crypto/cloudkms"
)

const (
	EnvPrefix = "FLOW"
)

var ConfigPath = []string{"flow.json"}

func Exit(code int, msg string) {
	fmt.Println(msg)
	os.Exit(code)
}

func Exitf(code int, msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(code)
}

func RandomSeed(n int) ([]byte, error) {
	seed := make([]byte, n)

	_, err := rand.Read(seed)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random seed: %v", err)
	}

	return seed, nil
}

var squareBracketRegex = regexp.MustCompile(`(?s)\[(.*)\]`)

// GcloudApplicationSignin signs in as an application user using gcloud command line tool
// currently assumes gcloud is already installed on the machine
// will by default pop a browser window to sign in
func GcloudApplicationSignin(resourceID string) error {
	googleAppCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if len(googleAppCreds) > 0 {
		return nil
	}

	kms, err := cloudkms.KeyFromResourceID(resourceID)
	if err != nil {
		return err
	}

	proj := kms.ProjectID
	if len(proj) == 0 {
		return fmt.Errorf("could not get GOOGLE_APPLICATION_CREDENTIALS, no google service account JSON provided but private key type is KMS")
	}

	loginCmd := exec.Command("gcloud", "auth", "application-default", "login", fmt.Sprintf("--project=%s", proj))

	output, err := loginCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Failed to run %q: %s\n", loginCmd.String(), err)
	}
	regexResult := squareBracketRegex.FindAllStringSubmatch(string(output), -1)
	// Should only be one value. Second index since first index contains the square brackets
	googleApplicationCreds := regexResult[0][1]

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", googleApplicationCreds)

	return nil
}
