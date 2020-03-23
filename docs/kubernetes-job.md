# Running load scenarios in Kubernetes

Gopherciser can run as a Kubernetes job, which executes the selected test from within a cluster and exposes metrics for Prometheus (an open-source systems monitoring and alerting toolkit). The job creates a pod that runs the test to completion or failure. For information on how to build a Docker image, see [Building a Docker image](./buildingdockerimage.md). 

This example uses an **externally** signed JSON Web Token (JWT) for simplicity and to avoid the requirement of having an authentication setup with predefined users. To set up your cluster this way, see [Running against QSEoK with JWT authentication](./random-qseok.md).

Do the following:

1. Create a Kubernetes ConfigMap to run a specific scenario:
   * Fetch the [test scenario](./examples/random-qseok.json).
   * Rename the test scenario file `testjob.json`.
   * Create the ConfigMap: `kubectl create configmap testconfig --from-file=testjob.json`
2. Create a secret for the JWT key to run a specific scenario:
   * Use the private key from step 1, Configuring the QSEoK deployment, in [Running against QSEoK with JWT authentication](./random-qseok.md).
   * The name of the key file (`gopherciser-key` by default) is specified in `connectionSettings.jwtsettings.keypath` and must be the same as specified in the `testjob.json` file created in the previous step.
   * Add the filename: `kubectl create secret generic gopherciser --from-file=gopherciser-key`
3. Run the test as a Kubernetes job: 
   * Fetch the [YAML test job](./examples/job.yaml).
   * Rename the YAML test job file `gopherciser-job.yaml`.
   * Create the Kubernetes job: `kubectl create -f kubernetes/gopherciser-job.yaml`
4. (Optional:) Create a service that exposes the metrics endpoint externally:
   * Fetch the [YAML test service](./examples/service.yaml).
   * Rename the YAML test service file `gopherciser-service.yaml`.
   * Create the service: `kubectl create -f kubernetes/gopherciser-service.yaml`
5. (Optional:) Modify the `gopherciser-job.yaml` file to run as a unique job so that several jobs can run in parallel: 
   ```
   metadata:
     generateName: k8-gopherciser #Instead of the name:, generates a unique job name based on k8-gopherciser
   ```
