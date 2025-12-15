# CloudPulse Commands (PowerShell)

### create the namespace

```bash
kubectl create namespace cloudpulse
```

### build images locally:

```bash
docker build -t cloudpulse-api:v1 -f apps/api/Dockerfile .
docker build -t cloudpulse-runner:v1 -f apps/runner/Dockerfile .
```

### database infrastructure

```bash
kubectl apply -f deployments/kubernetes/dynamodb.yaml -n cloudpulse
```

#### initialize tables

```bash
kubectl apply -f deployments/kubernetes/setup-tables-job.yaml -n cloudpulse
```

### deploy API via Helm

```bash
helm upgrade --install cloudpulse-api deployments/helm/cloudpulse-api -n cloudpulse --set image.repository=cloudpulse-api --set image.tag=v1 --set env.AWS_ENDPOINT="http://dynamodb-local:8000" --set env.TABLE_NAME_TARGETS="cloudpulse-targets-local" --set env.TABLE_NAME_RESULTS="cloudpulse-probe-results-local"
```

### deploy runner via manifest

```bash
kubectl apply -f deployments/kubernetes/runner.yaml -n cloudpulse
```

### port forwarding

```bash
kubectl port-forward svc/cloudpulse-api 8080:80 -n cloudpulse
```

### create new target:

```bash
Invoke-RestMethod -Uri "http://localhost:8080/targets" -Method Post -ContentType "application/json" -Body '{ "name": "Google", "url": "https://google.com" }'
```

### list all targets-local

```bash
Invoke-RestMethod -Uri "http://localhost:8080/targets" | Format-Table
```

### view all results

```bash
Invoke-RestMethod -Uri "http://localhost:8080/results" | Format-Table
```

### view results for given target

```bash
Invoke-RestMethod -Uri "http://localhost:8080/results/<target-id-here>" | Format-Table
```

### verify container names

```bash
kubectl get pod cloudpulse-api-65d58dc8c4-77wcf -n cloudpulse -o jsonpath="{.spec.containers[*].name}"
```

### debug container

```bash
kubectl debug -it cloudpulse-api-65d58dc8c4-77wcf -n cloudpulse --image=busybox --target=api
```

    ### Check Localhost Connectivity: Verify the API is listening on port 8080.
    ```bash
    wget -qO- http://localhost:8080/health
    ```

    ### confirm pod can resolve dynamodb-local to the underlying Service IP

```bash
	nslookup dynamodb-local

	telnet dynamodb-local 8000
		^]quit
		Console escape. Commands are:
		 l      go to line mode
		 c      go to character mode
		 z      suspend telnet
		 e      exit telnet
		e
		/ #
		/ # exit
```

### get k8s service

```bash
kubectl get svc -A
```

### clean out old manually-applied Deployment/service

```bash
kubectl -n cloudpulse delete deploy cloudpulse-api --ignore-not-found
kubectl -n cloudpulse delete svc cloudpulse-api --ignore-not-found
```

### search deployment

```bash
kubectl describe deployment -n cloudpulse | Select-String "Mounts:"
```

### Helm install/upgrade

```bash
helm upgrade --install cloudpulse-api deployments/helm/cloudpulse-api -n cloudpulse --create-namespace
```

### check helm release

```bash
helm list -n cloudpulse
```

### verify the Deployment, Service, and Pods:

```bash
kubectl -n cloudpulse get deploy,svc,pods
```

### Verification & Debugging

**Restart deployment (to speed up environment variable changes):**

```bash
kubectl rollout restart deploy cloudpulse-api -n cloudpulse
```

**Check logs for correct mode (DynamoDB vs In-Memory):**
_Look for "initializing dynamodb store"_

```bash
kubectl logs -l app.kubernetes.io/name=cloudpulse-api -n cloudpulse
```

**Port-Forward to a different port (if 8080 is stuck):**

```bash
kubectl port-forward svc/cloudpulse-api 8081:80 -n cloudpulse
```

### Metrics (Prometheus)

**Deploy:**

```bash
kubectl apply -f deployments/kubernetes/prometheus-configmap.yaml -n cloudpulse
kubectl apply -f deployments/kubernetes/prometheus.yaml -n cloudpulse
```

**Access Dashboard (localhost:9090):**

```bash
kubectl port-forward svc/prometheus 9090:9090 -n cloudpulse
```

### debug Service

```bash
kubectl describe svc -n $NS $SVC
kubectl get endpoints -n $NS $SVC
kubectl get endpointslices -n $NS -l kubernetes.io/service-name=$SVC
```
