# Example: Running against QSEoW with JWT authentication

This step-by-step example shows how to set up and run a randomworker scenario against a Qlik Sense® Enterprise for Windows (QSEoW) deployment with JSON Web Token (JWT) authentication enabled.

For an overview of how to use JWT authentication in QSEoW, see [Qlik Help: JWT authentication](https://help.qlik.com/en-US/sense/Subsystems/ManagementConsole/Content/Sense_QMC/JWT-authentication.htm).

**Note:** JWT authentication can be set up in different ways. The settings used in this example may not be valid for all QSEoW deployments.

## Requirements

The following is required for this procedure:

* A QSEoW deployment
* A server with Gopherciser installed (referred to as the "load client")
* An SSH key pair, generated on the server hosting the QSEoW deployment, to use with Gopherciser:
  * The private key is used to encrypt information passed in JWTs from the load client to the QSEoW deployment under test.
  * The public key, stored in an X.509 certificate, is used by the QSEoW deployment to decipher the information passed in the JWTs.

## Placing the private key on the load client

Place the file containing the private key used for JWT authentication in a location that can be accessed from the load client.

In this example, the file (`mock.pem`) is stored in the same folder as Gopherciser on the load client.

## Creating suitable access rules with enough tokens

Create access rules that allow allocation of licenses to the virtual users created by Gopherciser.

<details>

<summary>Example</summary>

In this example, a login access rule is used to allocate licenses to the virtual users created by Gopherciser. The rule allows users from a specific user directory ("anydir" in this example) to access the QSEoW deployment.

Do the following in the QSEoW deployment:

1. Open the Qlik® Management Console (QMC).
2. Go to **License management** > **Login access rules** > **Create new**.
3. Enter a name for the new login access rule in the **Name** field.
4. Enter the number of tokens allocated to the login access rule in the **Allocated tokens** field.
5. Select **Apply**. The **Create license rule** dialog appears with a default license rule selected. 
6. Select **Advanced** under **Properties** to display the code for the license rule.
7. Select **userDirectory** in the **name** drop-down.
8. Enter the name of the user directory, "anydir", in the empty field next to the **value** drop-down. 
9. Check that the code in the **Advanced** section is similar to the following: `((user.userDirectory="anydir"))`
10. Select **Apply** to create the login access rule.

For more information on how to create login access rules, see the [Qlik help](https://help.qlik.com/en-US/sense/Subsystems/ManagementConsole/Content/Sense_QMC/create-login-access.htm).

</details>

## Adding a virtual proxy for authentication of the virtual users

The next step is to create a virtual proxy to handle the authentication, session handling, and load balancing of the virtual users created by Gopherciser.

<details>

<summary>Example</summary>

Do the following in the QSEoW deployment:

1. Open the QMC.
2. Go to **Proxies**.
3. Select the proxy on the central node (**Central**) and then **Edit**.
4. Select **Virtual proxies** under **Associated items**.
5. Select **Add** > **Create new**.
6. Select **Authentication** and **Load balancing** under **Properties**.
7. Fill in the following in the **Identification** section:
   * **Description**: Enter a name for the new virtual proxy that will handle the virtual users ("virtualproxy" in this example).
   * **Prefix**: Enter the prefix to use for the new virtual proxy in the URL ("jwt" in this example).
   * **Session cookie header name**: Enter the name of the http header to use for the session cookie ("X-Qlik-Session-header" in this example).
8. Fill in the following in the **Authentication** section:
   * **Anonymous access mode**: Select **No anonymous user** in the drop-down.
   * **Authentication method**: Select **JWT** in the drop-down.
   * **JWT certificate**: Paste the JWT X.509 public key certificate in PEM format.
   * **JWT attribute for user ID**: Enter the JWT attribute name of the attribute that describes the user ID ("user" in this example).
   * **JWT attribute for user directory**: Enter the JWT attribute name of the attribute that describes the user directory ("directory" in this example).
9. Select **Add new server node** in the **Load balancing** section.
10. Select the engine nodes to load balance to and then select **Add**.
11. Select **Apply** to create the virtual proxy.

For information on how to create a virtual proxy, see the [Qlik help](https://help.qlik.com/en-US/sense/Subsystems/ManagementConsole/Content/Sense_QMC/create-virtual-proxy.htm).

</details>

## Importing and publishing the test apps

Import the test apps to the QSEoW deployment. Make sure to publish the apps, so that they are available to all users.

For information on how to publish apps, see the [Qlik help](https://help.qlik.com/en-US/sense/Subsystems/ManagementConsole/Content/Sense_QMC/publish-apps.htm).

## Modifying the sample test script

The sample script is available here: [General randomworker example with JWT authentication](examples/random-qseow.json)

<details>

<summary>Example</summary>

Do the following on the load client:

1. Download the sample test script.
2. Modify the following fields to match the QSEoW setup configured above:
   * `connectionSettings.server`: The hostname of the QSEoW deployment.
   * `connectionSettings.virtualproxy`: The prefix for the virtual proxy that handles the virtual users ("jwt" in this example).
   * `connectionSettings.jwtsettings.keypath`: The path to the private key (`mock.pem` in this example as the key is stored in the same folder as Gopherciser).
   * `connectionSettings.jwtsettings.claims`: The JWT claims as an escaped JSON string (`{\"user\":\"{{.UserName}}\",\"directory\":\"{{.Directory}}\"}` in this example). The claims must correspond to the **JWT attribute for user ID** and **JWT attribute for user directory** settings in the QMC.
   * `loginSettings.settings.directory`: The name of the user directory ("anydir" in this example). The directory name is used by the login access rule to allocate licenses.
   * `scenario.action: OpenApp.settings.randomapps`: The names of the test apps.
3. Save the changes to the script.

</details>

## Running the test script

Run the test script.

<details>

<summary>Example</summary>

Do the following on the load client:

1. Open a Command Prompt.
2. Execute the following command (the actual command differs depending on platform - the example below is based on Linux Bash):

```
./gopherciser execute -c random-qseow.json
```

</details>

The `settings.logs.filename` field in the test script specifies the name of and the path to the log file stored during the test execution.

## (Optional:) Viewing metrics in Grafana

To show continuous live [Prometheus](https://prometheus.io/) metrics during the execution, start Gopherciser with the following flag:
```
--metrics int
```
The exposed metrics include action metrics (such as response times per app and action), test warnings and test errors. 

The metrics are available at `http://localhost:port/metrics` during the test. Replace `port` with the port number specified by the `--metrics` flag.
