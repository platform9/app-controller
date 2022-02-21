## `fast-path`

A backend service that interacts with a Kubernetes cluster installed with knative serving components.

## Pre-requisites
To start the fast-path backend service, pre-requisites are:

1. Linux (64 bit)
2. A kubernetes cluster installed with [knative serving components](https://platform9.com/blog/how-to-set-up-knative-serving-on-kubernetes/)
3. MySQL Database to link to fast-path

## Configurations
The configurations for the service are set using `config.yaml`. Sample of `config.yaml` is present at [etc/config.yaml](etc/config.yaml)

This config.yaml should be placed at `/etc/pf9/fast-path/config.yaml`, it contains: 

```
# Path to kubeconfig file of the cluster that hosts knative
1. kubeconfig path

# Database name, username, password, URL, port. 
2. DB credentials

# auth0 JWKS URL, client id.
3. auth0 credentials

# constraints on maximum apps deploy count, replical count.
4. constraints (optional)
```

## Build fast-path

Clone the repository, navigate to the cloned repository and download the dependencies using `go mod download`. Before building, ensure the `config.yaml` is configured accordingly and placed at required location.

To build the fast-path binary, use the below command, fast-path binary is placed in `bin` directory

```sh
make build
```

## Run fast-path service

`fast-path` service can be run using binary and as a system service on linux machine.

### Using binary
To run fast-path through binary, follow the below command:
```sh
# Initialize and upgrade database.
./bin/fast-path migrate

# Start the fast-path service.
./bin/fast-path
```

### Using system service file
To run fast-path as a system service, service file [fastpath.service](fastpath.service) should be place at `/etc/systemd/system/` directory and fast-path binary at `/usr/bin/fast-path/` directory. To start the service follow the below commands:

```sh
# Start the fast-path service.
sudo systemctl start fastpath.service

# Check the status of fast-path service.
sudo systemctl status fastpath.service
```

Now, the fast-path service will be up an running.

* Logs for fast-path service can be found at `/var/log/pf9/fast-path/fast-path.log`

## `fast-path` APIs
To interace with fast-path service, fast-path APIs can be used and an auth0 token is required. 

```sh
# To get list of apps for a user.
curl --request GET --url 'http://127.0.0.1:6112/v1/apps'  --header "Authorization: Bearer ${AUTH0_IDTOKEN}" | jq .

# To describe an app by name.
curl --request GET --url 'http://127.0.0.1:6112/v1/apps/<name>'  --header "Authorization: Bearer ${AUTH0_IDTOKEN}" | jq .

# To create an app, where name is app name, image is container image of app, envs is environment variables with key:value pairs list, port is container port to access app.
curl --request POST --url 'http://127.0.0.1:6112/v1/apps'  --header "Authorization: Bearer ${AUTH0_IDTOKEN}" --data '{"name": "<appname>", "image": "<container image>", "envs": [{ "key":"<key>", "value":"<value>"}], "port": "<port>"}'

# To delete an app by name.
curl --request DELETE --url 'http://127.0.0.1:6112/v1/apps/<name>'  --header "Authorization: Bearer ${AUTH0_IDTOKEN}"
```

## Fetch auth0 token
Auth0 token can be fetched using auth0 apis. There are three major steps to get the [auth0 id token](https://auth0.com/docs/quickstart/native/device).

1. Request device code
2. Device activation
3. Request auth0 token

### **Pre-requisites**
* [auth0 native application](https://auth0.com/docs/get-started/auth0-overview/create-applications/native-apps)
* [auth0 device code setup](https://auth0.com/docs/quickstart/native/device#prerequisites)

#### [**Request device code**](https://auth0.com/docs/quickstart/native/device#request-device-code)

```sh
# Request device code 
curl --request POST \
  --url 'https://YOUR_DOMAIN/oauth/device/code' \
  --header 'content-type: application/x-www-form-urlencoded' \
  --data 'client_id=YOUR_CLIENT_ID' \
  --data 'scope=YOUR_SCOPE' 
```
- Basic sample scope is 'scope=profile email openid'

#### **Device activation**
Upon request for device code, the [sample device code](https://auth0.com/docs/quickstart/native/device#device-code-response) response will be:
```sh
{
  "device_code": "Ag_EE...ko1p",
  "user_code": "QTZL-MCBW",
  "verification_uri": "https://accounts.acmetest.org/activate",
  "verification_uri_complete": "https://accounts.acmetest.org/activate?user_code=QTZL-MCBW",
  "expires_in": 900,
  "interval": 5
}
```

Then open `verification_url_complete` in browser, obtained from device code response to complete the device activation. 

#### [**Request auth0 token**](https://auth0.com/docs/quickstart/native/device#example-request-token-post-to-token-url)

Once device activation is successful, then request for auth0 token using below command.
```sh
# Use the device code received from device code response
curl --request POST \
  --url 'https://YOUR_DOMAIN/oauth/token' \
  --header 'content-type: application/x-www-form-urlencoded' \
  --data grant_type=urn:ietf:params:oauth:grant-type:device_code \
  --data device_code=YOUR_DEVICE_CODE \
  --data 'client_id=YOUR_CLIENT_ID'
```

The received token will containe both access_token, id_token. We use auth0 `id_token` to authorize the user through fast-path. To access the fast-path APIs seamlessly export the auth0 `id_token`. 

```sh
# Replace the <id_token> with the id_token value received from request auth0 token.
export AUTH0_IDTOKEN="<id_token>"
```