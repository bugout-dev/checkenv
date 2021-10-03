# checkenv

Don't get surprised by your environment variables.

## Rationale

At Bugout, we configure our applications using environment variables. This follows the [Twelve Factor
App methodology](https://12factor.net/config) and generally serves us well.

Still, there are many ways to pass environment variables to an application process. For example, we
store environment variables on
[AWS Systems Manager Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
and on [Google Cloud Secret Manager](https://cloud.google.com/secret-manager). When we deploy applications
to our servers, the first step is to download these environment variables from the cloud store and
write them to environment files. We then pass these environment files to the [`systemd`](https://en.wikipedia.org/wiki/Systemd)
services responsible for running the applications using the `EnvironmentFile` parameter.

In our development environments, we have files that are very similar to the `systemd` environment files,
except they explicitly `export` the environment variables. We source these environment files before we
run the development versions of our applications.

It can sometimes be difficult to understand:
1. Whether all the environment variables we *expect* to be defined in production actually have been.
2. What the particular value of a production environment actually is.
3. What the differences are between our expectations and the actual environment variables in a running
application process.

We are building and maintaining `checkenv` to make it easier for us to diagnose and fix issues with
application configuration via environment variables. We stand in solidarity with anyone else who
experiences these kinds of issues. We are doing this for you, too! :)

## Architecture

Because environment variables can come from different providers - an active shell, an environment file,
a cloud configuration store - `checkenv` allows environment providers to be registered at build time
using build tags.

This means that you can write a custom environment provider for `checkenv` and build a custom `checkenv`
binary which supports your needs.

There is currently no need to support runtime plugins. Since doing so would make this program a lot
more complicated, we have decided to forego runtime plugin functionality for now.
