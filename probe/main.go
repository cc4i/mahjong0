package main

import (
	"bufio"
	"flag"
	"io"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

// probeHttp for http style
func probeHttp() {
	//TODO add implementation
}

// probeCmd for command style
func probeCmd(command string, periodSeconds int, successThreshold int, failureThreshold int) error {

	failure := 0
	success := 0

	for {
		cmd := exec.Command(command)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()
		err := cmd.Start()
		if err != nil {
			log.Error(err)
			failure++
		}
		tail(stdout, false)
		tail(stderr, true)
		err = cmd.Wait()
		if err != nil {
			failure++
		} else {
			success++
		}

		if failure >= failureThreshold && failureThreshold != -1 {
			log.Infof("Exit at failure : failure=%d, success=%d", failure, success)
			log.Error(err)
			return err
		}
		if success >= successThreshold && successThreshold != -1 {
			log.Infof("exit at success : failure=%d, success=%d", failure, success)
			log.Info("The service is ready!")
			return nil
		}
		if !cmd.ProcessState.Exited() {
			err = cmd.Process.Kill()
			if err != nil {
				log.Errorf("!!!Not exited and failed to kill!!! : failure=%d, success=%d : error=%s", failure, success, err)
			}
		}

		log.Infof("!!!In the Loop!!!: failure=%d, success=%d", failure, success)

		time.Sleep(time.Duration(periodSeconds) * time.Second)
	}

}

// tail read output from stream
func tail(reader io.ReadCloser, isErr bool) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		buf := scanner.Bytes()
		if isErr {
			log.Errorf("%s", buf)
		} else {
			log.Infof("%s", buf)
		}

	}

}

func main() {

	command := flag.String("command", "", "Probe command")
	initialDelaySeconds := flag.Int("initialDelaySeconds", 0, "Initial delay(s) for probe command")
	periodSeconds := flag.Int("periodSeconds", 0, "Period(s) for probe command")
	timeoutSeconds := flag.Int("timeoutSeconds", 0, "Timeout(s) for probe command")
	successThreshold := flag.Int("successThreshold", 0, "Maximum count of success")
	failureThreshold := flag.Int("failureThreshold", 0, "Maximum count of failure")
	flag.Parse()
	if flag.NFlag() != 6 {
		log.Fatal("Input parameters weren't expected")
	}
	log.Println("The probe is kicking start ...")
	var signal = make(chan string, 1)

	go func() {
		time.Sleep(time.Duration(*initialDelaySeconds) * time.Second)
		err := probeCmd(*command, *periodSeconds, *successThreshold, *failureThreshold)
		if err != nil {
			log.Error(err)
			signal <- "fail"
		} else {
			signal <- "success"
		}
	}()

	select {
	case res := <-signal:
		if res == "fail" {
			log.Fatalf("Probe - %s was failed", *command)
		} else {
			log.Infof("Probe - %s was success", *command)
		}
	case <-time.After(time.Duration(*timeoutSeconds) * time.Second):
		log.Fatalf("Probe - %s is timeout", *command)
	}
	log.Println("Done")
}
