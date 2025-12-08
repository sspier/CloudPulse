# CloudPulse Commands

- build: docker build -f apps/api/Dockerfile .
- run: docker compose -f deployments/compose/docker-compose.local.yml up --build
- see health: curl -v http://localhost:8080/health
- see logs for container: 'docker ps' then 'docker logs cloudpulse-api-1'
- create a target: curl -v http://localhost:8080/targets -H "Content-Type: application/json" -d '{ "name": "My Blog", "url": "https://example.com"}'
- list all targets: curl -v http://localhost:8080/targets
- get latest results for each target: curl -v http://localhost:8080/results
- get full history for one target: curl -v http://localhost:8080/results/abc123


- Build the image locally: CloudPulse\apps\api> docker build -t cloudpulse-api:dev .
- Update deployment: CloudPulse\deployments\kubernetes> kubectl apply -f .\cloudpulse-api-deployment.yaml
- Update deployment: CloudPulse\deployments\kubernetes> kubectl apply -f .\cloudpulse-api-service.yaml
- See pods and their state: kubectl -n cloudpulse get pods
- See details for a specific pod: kubectl -n cloudpulse describe pod cloudpulse-api-xxxxxxxxxx-xxxxx
- Port-forwarding: 
  - Create a network pipe from the local machine to a pod or service inside the cluster:
	- kubectl -n cloudpulse port-forward svc/cloudpulse-api 8080:80
  - Create a pipe directly to a pod:
    - kubectl port-forward pod/cloudpulse-api-xxxxx 8080:8080

- Clean out old manually-applied Deployment/service
  - kubectl -n cloudpulse delete deploy cloudpulse-api --ignore-not-found
  - kubectl -n cloudpulse delete svc cloudpulse-api --ignore-not-found
	
- Helm	
  - Install / upgrade the Helm release: helm upgrade --install cloudpulse-api deployments/helm/cloudpulse-api -n cloudpulse --create-namespace
  - Check the release: helm list -n cloudpulse
  - Verify the Deployment, Service, and Pods: kubectl -n cloudpulse get deploy,svc,pods


