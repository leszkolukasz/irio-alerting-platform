# Alerting platform

## Setup

Flake describes all necessary tools and their versions. To enter the development environment, run:

```bash
nix --extra-experimental-features "nix-command flakes" develop
```

or install the dependencies manually.

Direnv is configured to automatically enter the environment when you `cd` into the project directory.

## Environment variables

File `.envrc.example` contains a list of required environment variables. Copy it to `.envrc` and fill in the values. Then either source it or let direnv load it automatically.

Each microservice may need additional environment variables defined in `.env` files in their respective directories.

## Code formatting and linting

Makefile contains various commands for building, linting and formatting the code.

## Running the application

This is a monorepo. See the `frontend` and `backend/*` directories for instructions on running the respective parts of the application.
