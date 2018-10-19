package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"code.cloudfoundry.org/cli/plugin"
)

type MultiCmd struct{}

func (c *MultiCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "osb-local",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 20,
		},
		Commands: []plugin.Command{
			{
				Name:     "osb",
				HelpText: "Add and remove service brokers in cf dev",
				UsageDetails: plugin.Usage{
					Usage: "osb - add and stop remove service brokers in cf dev\n\nUsage:\n    cf osb [command] \n\n    Available Commands:\n        add SERVICE_BROKER URL       add new service broker from remote located container\n        remove SERVICE_BROKER URL    remove added service broker using remote located container\n\n    Prerequesites:\n        Installed bosh cli on your machine and you have to run `cf dev start` first.\n\n    Usage Example:\n        cf osb add myservicebroker https://my-file.server/myosb.tgz",
				},
			},
		},
	}
}

func main() {

	plugin.Start(new(MultiCmd))
}

func (c *MultiCmd) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "osb" {
		if args[1] == "add" {
			c.Add(args[2:])
		} else if args[1] == "remove" {
			c.Remove(args[2:])
		}
	}
}

func (c *MultiCmd) Add(args []string) {
	fmt.Println("Adding new service broker to cf dev...")
	script := "./osb-add.sh"
	c.ExecuteScript(script, addSh, args)
}

func (c *MultiCmd) Remove(args []string) {
	fmt.Println("Removing service broker from cf dev...")
	script := "./osb-remove.sh"
	c.ExecuteScript(script, removeSh, args)
}

func (c *MultiCmd) ExecuteScript(scriptName string, script string, args []string) {
	c.CreateScript(scriptName, script)
	cmd := exec.Command(scriptName, args...)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		log.Printf(m)
	}

	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		m := errScanner.Text()
		log.Printf(m)
	}

	cmd.Wait()

	if _, err := os.Stat(scriptName); !os.IsNotExist(err) {
		err := os.Remove(scriptName)

		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (c *MultiCmd) CreateScript(script string, code string) {
	data := []byte(code)
	err := ioutil.WriteFile(script, data, 0777)
	check(err)
}

var addSh = `#!/bin/bash

set -eu

echo "Starting with the adding..."

if [[ -z "$1" ]]; then
  exit 1
fi

SERVICE_BROKER=$1


if [[ -z "$2" ]]; then
exit 1
fi

URL=$2

USERNAME="admin"
PASSWORD="secret"

eval "$(cf dev bosh env)"

export BOSH_ENVIRONMENT
export BOSH_CLIENT
export BOSH_CLIENT_SECRET
export BOSH_CA_CERT
export BOSH_GW_HOST
export BOSH_GW_PRIVATE_KEY
export BOSH_GW_USER

echo "Setup environment finished"

cf api https://api.dev.cfdev.sh  --skip-ssl-validation
cf auth admin admin

cf create-space -o system service-brokers
cf t -o system -s service-brokers

echo "Login to cf dev finished"

FILENAME=$(basename "$URL")
if [[ -f "$FILENAME" ]]; then
  rm $FILENAME
fi
wget $URL --no-cache -q 2>&1

if [[ -d "service-broker" ]]; then
  rm -R service-broker
fi

mkdir service-broker
tar -xzf $FILENAME -C service-broker

echo "Service broker ready to add"

pushd ./service-broker > /dev/null
SERVICE=$(sh ./deploy.sh $SERVICE_BROKER $USERNAME $PASSWORD 2>&1 | tee -a osb.log | tail -n 1)
popd > /dev/null

echo "Done with deploying service broker"

cf create-service-broker $SERVICE_BROKER $USERNAME $PASSWORD "https://$SERVICE_BROKER.dev.cfdev.sh"
cf enable-service-access $SERVICE

echo "Added service broker"

rm -R $FILENAME
rm -R ./service-broker

echo "All cleaned up"
`

var removeSh = `#!/bin/bash

set -eu

echo "Starting with the removing..."

SERVICE_BROKER=$1
URL=$2

echo "Setup environment finished"

FILENAME=$(basename "$URL")
if [[ -f "$FILENAME" ]]; then
  rm $FILENAME
fi
wget $URL --no-cache -q 2>&1


if [[ ! -d "service-broker" ]]; then
  mkdir service-broker
  tar -xzf $FILENAME -C service-broker
fi

echo "Service broker ready to remove"

cf api https://api.dev.cfdev.sh  --skip-ssl-validation
cf auth admin admin

cf create-space -o system service-brokers
cf t -o system -s service-brokers

echo "Login to cf dev finished"

cf delete-service-broker $SERVICE_BROKER -f

echo "Removed service broker"

sh ./service-broker/remove.sh $SERVICE_BROKER

echo "Deleted service broker"

rm $FILENAME
rm -R ./service-broker

echo "All cleaned up"
`
