# hyper-selenium

⚠️ ⚠️ **Go Noob Warning:** This is my first ever Go project. Please expect to see a lot of stupid noobish unidiomatic Go code in here. PRs welcome. ⚠️ ⚠️

## Architecture

<img src="https://docs.google.com/drawings/d/e/2PACX-1vSmoU3tAQIhgiMLSLD0Ut-XUBv41VqJdVbCElUL3bEAusVq3QaoORLnTXQGVsxzqx9X6ejYj29KSCCt/pub?w=899&amp;h=223">

- **Local machine:** The machine running the E2E test.

  - Contains the test scripts.
  - Not powerful enough to run all of these tests at the same time.

- **Docker container:** This container runs Selenium.

  - Can be spawned on-demand. e.g. 20 containers are spawned to run 20 tests simultaneously.
  - For cost-effectiveness, should be billed per-second. [Hyper.sh](https://hyper.sh/) offers this.
  - It might not have its own IP address, so a secure tunnel must be employed.

- **SSH server:** This server allows the Local machine and Docker container to communicate with each other.

TODO talk about agent and client.

## Development

### Agent

Building Docker container

```
docker build -t hyper-selenium-agent .
```

Running

```
docker run -ti --rm hyper-selenium-agent
```

### Local development workflow

Running a central ssh server:

```
docker run -d -p 2222:22 --name hyper-selenium-sshd rastasheep/ubuntu-sshd:18.04
```

Compiling and running the agent:

```
./scripts/run-agent-dev.sh --ssh-remote=192.168.2.62:2222 --id=test
```

Running the client:

```
go run ./cmd/hyper-selenium-client/main.go --ssh-remote=192.168.2.62:2222 bash
```
