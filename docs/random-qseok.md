# Example: Running against QSEoK with JWT authentication

This step-by-step example shows how to set up and run a randomworker scenario against a Qlik Sense® Enterprise on Kubernetes (QSEoK) deployment with **external** JSON Web Token (JWT) authentication enabled.

For an overview of how to use JWT authentication in Qlik Sense, see [Qlik Help: JWT authentication](https://help.qlik.com/en-US/sense/Subsystems/ManagementConsole/Content/Sense_QMC/JWT-authentication.htm).

**Note:** JWT authentication can be set up in different ways. The settings used in this example may not be valid for all QSEoK deployments.

## Requirements

The following is required for this procedure:

* A Kubernetes cluster with a QSEoK deployment
* A Secure Shell (SSH) client (for example, Cygwin or PuTTY) for communication with the QSEoK deployment
* A server with Gopherciser installed (referred to as the "load client")
* An SSH key pair, generated on the Master Kubernetes node for the QSEoK deployment, to use with Gopherciser:
  * The private key is used to encrypt information passed in JWTs from the load client to the QSEoK deployment under test.
  * The public key is used by the QSEoK deployment to decipher the information passed in the JWTs.

## Generating the SSH key pair

First, generate the keys needed for Gopherciser.

<details>

<summary>Example</summary>

**Note:** This example requires OpenSSL to be installed.

Do the following in the QSEoK deployment:

1. Generate the private key and save it to a file (`mock.pem`):
```
openssl genrsa -out mock.pem 4096
```

2. Generate the public key and save it to a file (`mockpublic.pem`):
```
openssl rsa -in mock.pem -pubout -out mockpublic.pem
```

3. Store the private key in a location that can be accessed by the load client. In this example, the file (`mock.pem`) is stored in the same folder as Gopherciser on the load client.

</details>

## Configuring the QSEoK deployment

The next step is to configure the QSEoK deployment.

### Enabling JWT authentication

Enable JWT authentication, so that the load client can communicate with the QSEoK deployment.

<details>

<summary>Example</summary>

Do the following in the QSEoK deployment:

* Enable JWT authentication using the `identity-providers` module in the "values" file.

  **Warning:** The `identity-providers` configuration below is not intended for production use and the provided example may be subject to change without notice. In addition, since no cookie is stored, every request is evaluated for authentication.

  **Note:** The contents of the `kid` (key ID) field must match the contents of the corresponding field in the test script (see `connectionSettings.jwtsettings.jwtheader` below).

```
identity-providers:
  secrets:
    idpConfigs:
      - issuerConfig:
          issuer: "https://qlik.api.internal"
        primary: false
        realm: "custom"
        hostname: "my-server-url"
        staticKeys:
        - kid: "my-key-identifier"
          pem: |-
            -----BEGIN PUBLIC KEY-----
            <INSERT PUBLIC KEY HERE>
            -----END PUBLIC KEY-----
```

</details>

### Configuring certificate 

By default, a self-signed certificate is used when running Gopherciser. 

<details>

<summary>Example</summary>

To use a self-signed certificate, do the following:

* Set `connectionSettings.allowuntrusted` to `true` in your test script (as it is set to `false` by default).

To use a proper certificate instead, do the following:

1. Add the certificate and its private key using the `elastic-infra` module in the "values" file (alternatively, store the key as a secret in the Kubernetes cluster and point to it from the "values" file).

   **Note:** The private key below is not the same as the public key in the `identity-providers` module, but rather the valid domain certificate information.

```
elastic-infra:
  tlsSecret:
    enabled: true
    crt: |
      -----BEGIN CERTIFICATE-----
      <INSERT CERTIFICATES HERE>
      -----END CERTIFICATE-----
    key: |
      -----BEGIN PRIVATE KEY-----
      <INSERT PRIVATE KEY HERE>
      -----END PRIVATE KEY-----
```

2. Set `connectionSettings.security` to `true` in your test script (as it is set to `false` by default).

</details>

## Licensing

Configure the licensing in the QSEoK deployment, so that the virtual users created by Gopherciser are allocated licenses when they log in.

<details>
 
<summary>Example</summary>

Do the following in the QSEoK deployment:

1. Open the Management Console.
2. Select **License / User Allocation** under **Governance**.
3. Configure the licensing for the virtual users.

</details>

## Importing and publishing the test apps

Import the test apps to the QSEoK deployment. Make sure to publish the apps and make all sheets in them public, so that they are available to all users.

## Modifying the sample test script

The sample test script is available here: [Random worker example with external JWT](./examples/random-qseok.json)

<details>
 
<summary>Example</summary>

Do the following on the load client:

1. Download the sample test script.
2. Modify the following fields to match the QSEoK setup configured above:

   * `connectionSettings.server`: The hostname of the QSEoK deployment.
   * `connectionSettings.jwtsettings.keypath`: The path to the private key (`mock.pem` in this example as the key is stored in the same folder as Gopherciser).
   * `connectionSettings.jwtsettings.jwtheader`: The JWT headers as an escaped JSON string.
   * `connectionSettings.jwtsettings.claims`: The JWT claims as an escaped JSON string.
   * `loginSettings.settings.directory`: The name of the user directory.
   * `scenario.action: OpenApp.settings.appguid`: The GUID of the test app.

3. Save the changes to the script.

</details>

## Running the test script

Run the test script on the load client by executing the following command (the actual command differs depending on platform - the example below is based on Linux Bash):

```
./gopherciser execute -c random-qseok.json
```

The `settings.logs.filename` field in the test script specifies the name of and the path to the log file stored during the test execution.

## (Optional:) Viewing metrics in Grafana

To show continuous live [Prometheus](https://prometheus.io/) metrics during the execution, start Gopherciser with the following flag:
```
--metrics int
```
The exposed metrics include action metrics (such as response times per app and action), test warnings and test errors. 

The metrics are available at `http://localhost:port/metrics` during the test. Replace `port` with the port number specified by the `--metrics` flag.
