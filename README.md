# hyper-selenium

## The story

This is the story of our E2E test infrastructure.

### Part 1: Selenium running on CircleCI

The simplest E2E solution is to run Selenium locally on the CI. However, a single CI container is only capable of running a single test at a time. Running tests in parallel makes it very unstable.

To get more parallelism, we would need to rent more containers, each costs $50 per month.

### Part 2: Selenium running on DigitalOcean machines

To solve this problem, we launched 5 droplets on DigitalOcean.
Each droplet costs $10 per month and can handle a single running Selenium instance. This allows us to run E2E tests 5x faster.

However, there are significantly more moving parts that we have to manage.

- We had to self-manage a cluster of Selenium servers.
- We had to create a system that allocates tasks to an available container.

### Part 3: Selenium running on Hyper.sh

## Workflow

### Agent

Building Docker container

```
docker build -t hyper-selenium-agent .
```

Running

```
docker run -ti --rm hyper-selenium-agent
```

Development workflow (must build container first)

```
./scripts/run-agent-dev.sh
```