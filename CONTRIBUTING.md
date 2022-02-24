# Contributing to `fast-path` #

Thanks for your help improving the fast-path!

## Getting Help ##

If you have questions around `fast-path`, please contact the platform9 support or if you encounter any problem using it do raise an [Issue](https://github.com/platform9/fast-path/issues), you can reach us via email at [support@platform9.com](support@platform9.com).


## Workflow ##

Contribute to `fast-path` by following the guidelines below:
- [Clone](https://github.com/platform9/fast-path.git)
- Make sure all [fast-path prerequisites](https://github.com/platform9/fast-path#pre-requisites) are met
- [Auth0 prerequisites](https://github.com/platform9/fast-path#pre-requisites-1)
- [Setup local environment](https://github.com/platform9/fast-path#configurations) 
- Work on the cloned repository
- Open a pull request to [merge into fast-path repository](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request)


## Building ## 

Follow these steps to build from source and deploy:

1. [Build fast-path binary](https://github.com/platform9/fast-path#build-fast-path)
2. [For DB Schema changes or first time builds](https://github.com/platform9/fast-path#for-db-schema-changes-or-first-time-builds).
3. Run fast-path either [using binary](https://github.com/platform9/fast-path#using-binary) or [using system service file (preferred)](https://github.com/platform9/fast-path#using-system-service-file-preferred).
4. Logs for fast-path service can be found at `/var/log/pf9/fast-path/fast-path.log`

## Running the unit tests and manual test ##
1. To run unit tests:
```sh
make test
```
2. When submitting the pull request, perform manual tests using [fast-path APIs](https://github.com/platform9/fast-path#fast-path-apis) and attach snapshots or output files.
```sh
# List apps.
1. getApp API
# Deploy an app.
2. createApp API
# Describe an app.
3. getAppByName API
# Delete an app.
4. deleteApp API
```

## Committing ###

Please follow the Pull request template, before raising a PR, so reviewers will gain a deeper understanding as they review. If an outstanding issue has been fixed, please include the Fixes Issue # in your commit message.
