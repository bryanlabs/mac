package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/alecthomas/kingpin.v2"
	ini "gopkg.in/ini.v1"
)

var (
	// Flag 'profiles' is a comma seperated list of aws named profiles. EG: mac -a 'prod,stage,dev'. This flag is required.
	profiles = kingpin.Flag("profiles", "String of profile names.").Required().Short('p').String()

	// Flag 'maxRunners' is a number of jobs to run in parallel. EG: mac -n '4'. This flag is optional, 4 is default.
	maxRunners = kingpin.Flag("maxRunners", "Max Number of Parallel runners.").Default("4").Short('n').Int()

	// Argument 'cmd' is the command to execute. EG: mac -a 'dev,prod' 'aws sts get-caller-identity'
	cmd = kingpin.Arg("cmd", "The Command to execute.").Required().String()

	// Create a waitgroup to handle go routines.
	wg sync.WaitGroup
)

func main() {
	// Application version.
	kingpin.Version("2018.09.08")

	// Validate the commandline arguments and flags.
	kingpin.Parse()

	// Convert the provided account numbers to a slice.
	profiles := strings.Split(*profiles, ",")

	// Start a waitgroup based on the number of accounts provided.
	wg.Add(len(profiles))

	// Create a channel to hold jobs for runners.
	c := make(chan int, *maxRunners)

	// Run a job for each account in the slice.
	for _, profile := range profiles {
		go mac(profile, *cmd, c)
	}

	// wait for mac to finish processing all jobs before exiting.
	wg.Wait()

}

// mac will run a command under the specified profile.
func mac(profile string, cmd string, c chan int) {
	// Fill the channel buffer with random data.
	c <- 1

	// Run command in the correct context.
	slice := strings.Split(cmd, " ")
	cmdName := slice[0]
	cmdArgs := slice[1:]
	matchedProfile := getMatchedProfile(profile)
	macRun(matchedProfile, cmdName, cmdArgs)

	// Remove the job from the channel.
	<-c

	// Send Done when nothing remains on the channel.
	wg.Done()

}

// Run the command in the correct context(s).
func macRun(profile string, cmdName string, cmdArgs []string) {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Env = append(os.Environ(), "AWS_PROFILE="+profile, "AWS_SDK_LOAD_CONFIG=1")
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("Profile: %v, Error Creating Cmd: %v\n", profile, err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		var outputslice []string
		var stdout string
		for scanner.Scan() {
			outputslice = append(outputslice, "\n", scanner.Text())
			stdout = strings.Join(outputslice, "")
		}
		fmt.Printf("Profile: %s", profile)
		fmt.Printf("%s\n", stdout)
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Printf("Profile: %v, Error Starting Cmd: %v\n", profile, err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Profile: %v, Error waiting for Cmd: %v\n", profile, err)
		os.Exit(1)
	}

}

// getAWSConfig will return the filename for the users awsconfig.
func getAWSConfig() (awsconfig string) {
	switch flavor := runtime.GOOS; flavor {
	case "windows":
		return os.Getenv("userprofile") + "/.aws/config"
	case "linux":
		return os.Getenv("HOME") + "/.aws/config"
	default:
		log.Fatalf("Can't find aws config. %v is an Unsupported OS, please consider contributing to add this feature.", runtime.GOOS)
	}
	return
}

// getNamedProfile will find the named profile that matches the string provided.
func getMatchedProfile(profilestr string) (profile string) {

	awsconfig := getAWSConfig()
	cfg, err := ini.Load(awsconfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error Loading awsConfig: ", err)
		os.Exit(1)
	}

	// Find the profile.
	for _, section := range cfg.Sections() {
		if section.HasKey("role_arn") {

			slice := strings.Split(section.Name(), " ")
			namedProfile := slice[1]
			if namedProfile == profilestr {
				words := section.Name()
				slice := strings.Split(words, " ")
				profile = slice[1]
			}
		}
	}
	return
}
