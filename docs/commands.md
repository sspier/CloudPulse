# CloudPulse Commands (PowerShell)


### create the namespace
kubectl create namespace cloudpulse

### build images locally: 
docker build -t cloudpulse-api:v1 -f apps/api/Dockerfile .
docker build -t cloudpulse-runner:v1 -f apps/runner/Dockerfile .

### database infrastructure
kubectl apply -f deployments/kubernetes/dynamodb.yaml -n cloudpulse

#### initialize tables
kubectl apply -f deployments/kubernetes/setup-tables-job.yaml -n cloudpulse

### deploy API via Helm
helm upgrade --install cloudpulse-api deployments/helm/cloudpulse-api -n cloudpulse --set image.repository=cloudpulse-api --set image.tag=v1 --set env.AWS_ENDPOINT="http://dynamodb-local:8000" --set env.TABLE_NAME_TARGETS="cloudpulse-targets-local" --set env.TABLE_NAME_RESULTS="cloudpulse-probe-results-local"

### deploy runner via manifest
kubectl apply -f deployments/kubernetes/runner.yaml -n cloudpulse

### port forwarding
kubectl port-forward svc/cloudpulse-api 8080:80 -n cloudpulse

### create new target:
Invoke-RestMethod -Uri "http://localhost:8080/targets" -Method Post -ContentType "application/json" -Body '{ "name": "Google", "url": "https://google.com" }'	

### list all targets-local
Invoke-RestMethod -Uri "http://localhost:8080/targets" | Format-Table

### view all results
Invoke-RestMethod -Uri "http://localhost:8080/results" | Format-Table

### view results for given 
Invoke-RestMethod -Uri "http://localhost:8080/results/<target-id-here>" | Format-Table

### verify container names 
kubectl get pod cloudpulse-api-65d58dc8c4-77wcf -n cloudpulse -o jsonpath="{.spec.containers[*].name}"

### debug container
kubectl debug -it cloudpulse-api-65d58dc8c4-77wcf -n cloudpulse --image=busybox --target=api

	### Check Localhost Connectivity: Verify the API is listening on port 8080.
	wget -qO- http://localhost:8080/health
	
	### confirm pod can resolve dynamodb-local to the underlying Service IP
	nslookup dynamodb-local

	### 
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
	
	
	

- Clean out old manually-applied Deployment/service
  - kubectl -n cloudpulse delete deploy cloudpulse-api --ignore-not-found
  - kubectl -n cloudpulse delete svc cloudpulse-api --ignore-not-found
	
- Helm	
  - Install / upgrade the Helm release: helm upgrade --install cloudpulse-api deployments/helm/cloudpulse-api -n cloudpulse --create-namespace
  - Check the release: helm list -n cloudpulse
  - Verify the Deployment, Service, and Pods: kubectl -n cloudpulse get deploy,svc,pods


netstat -ano | findstr :8080

tasklist /FI "PID eq 13324"
tasklist /FI "PID eq 13836"


taskkill /PID <PID> /F



kubectl get svc -A

