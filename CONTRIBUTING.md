# Contributing to `app-controller` #

Thanks for your help improving the app-controller!

## Getting Help ##

If you have questions around `app-controller`, please contact the platform9 support or if you encounter any problem using it do raise an [Issue](https://github.com/platform9/app-controller/issues), you can reach us via email at [support@platform9.com](support@platform9.com).


## Workflow ##

Contribute to `app-controller` by following the guidelines below:
- [Clone](https://github.com/platform9/app-controller.git)
- Make sure all [app-controller prerequisites](https://github.com/platform9/app-controller#pre-requisites) are met
- [Auth0 prerequisites](https://github.com/platform9/app-controller#pre-requisites-1)
- [Setup local environment](https://github.com/platform9/app-controller#configurations) 
- Work on the cloned repository
- Open a pull request to [merge into app-controller repository](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request)


## Building ## 

Follow these steps to build from source and deploy:

1. [Build app-controller binary](https://github.com/platform9/app-controller#build-app-controller)
2. [For DB Schema changes or first time builds](https://github.com/platform9/app-controller#for-db-schema-changes-or-first-time-builds).
3. Run app-controller either [using binary](https://github.com/platform9/app-controller#using-binary) or [using system service file (preferred)](https://github.com/platform9/app-controller#using-system-service-file-preferred).
4. Logs for app-controller service can be found at `/var/log/pf9/app-controller/app-controller.log`

## Running the unit tests and manual test ##
1. To run unit tests:
```sh
make test
```
2. When submitting the pull request, perform manual tests using [app-controller APIs](https://github.com/platform9/app-controller#app-controller-apis) and attach snapshots or output files.
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
