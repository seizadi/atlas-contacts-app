# atlas-contacts-app

_This generated README.md file loosely follows a [popular template](https://gist.github.com/PurpleBooth/109311bb0361f32d87a2)._

One paragraph of project description goes here.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### Prerequisites

Install go dep

``` sh
go get -u github.com/golang/dep/cmd/dep
```

Install atlas-db-controller for managing database resources:

  * To create atlas-db custom resource and controller follow [link](https://github.com/infobloxopen/atlas-db/blob/master/README.md).
  * To create database instance, database and schema/tables resources locally to be used by contacts app, modify `contacts-localdb.yaml` following specifications mentioned in [link](https://github.com/infobloxopen/atlas-db/blob/master/README.md) and run:
  ```
    make db-up
  ```
  * To use RDS database instance, database and schema/tables resources used by contacts app, modify `contacts-rds.yaml` and run"
  ```sh
    kubectl create -f ./deploy/contacts-rds.yaml
  ```

### Local development setup

Please note that you should have the following ports opened on you local workstation: `:8080 :8081 :9090 :5432`.
If they are busy - please change them via corresponding parameters of `gateway` and `server` binaries or postgres container run.

Run PostgresDB:

```sh
docker run --name contacts-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=contacts -p 5432:5432 -d postgres:9.4
```

Table creation should be done manually by running the migrations scripts. Scripts can be found at `./db/migrations/`

Create vendor directory with required golang packages
``` sh
make vendor
```

Run App server:

``` sh
go run ./cmd/server/*.go -db "host=localhost port=5432 user=postgres password=postgres sslmode=disable dbname=contacts"
```

#### Try atlas-contacts-app

For Multi-Account environment, Authorization token (Bearer) is required. You can generate it using https://jwt.io/ with following Payload:
```
{
  "AccountID": YourAccountID
}
```

Example:
```
{
  "AccountID": 1
}
```
Bearer
``` sh
export JWT="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2NvdW50SUQiOjF9.GsXyFDDARjXe1t9DPo2LIBKHEal3O7t3vLI3edA7dGU"
```

Request examples:
``` sh
curl -H "Authorization: Bearer $JWT" \
http://localhost:8080/v1/contacts -d '{"first_name": "Mike", "primary_email": "mike@example.com"}'
```

``` sh
curl -H "Authorization: Bearer $JWT" \
http://localhost:8080/v1/contacts -d \
'{"first_name": "Robert", "primary_email": "robert@example.com", "nicknames": ["bob", "robbie"]}'
```

``` sh
curl -H "Authorization: Bearer $JWT" \
http://localhost:8080/v1/contacts?_filter='first_name=="Mike"'
```
Note, that `JWT` should contain AccountID field.

#### Build docker images

``` sh
make
```
Will be created docker images 'infoblox/contacts-gateway' and 'infoblox/contacts-server'.

If this process finished with errors it's likely that docker doesn't allow to mount host directory in its container.
Therefore you are proposed to run `su -c "setenforce 0"` command to fix this issue.

### Local Kubernetes setup

##### Prerequisites

Make sure nginx is deployed in your K8s. Otherwise you can deploy it using

``` sh
make nginx-up
```

##### Deployment
To deploy atlas-contacts-app use

``` sh
make up
```
Will be used latest Docker Hub images: 'infoblox/contacts-gateway:latest', 'infoblox/contacts-server:latest'.

To deploy authN stub, clone atlas-stubs repo (https://github.com/infobloxopen/atlas-stubs.git) and then execute deployment script inside authn-stub package or:

``` sh
curl https://raw.githubusercontent.com/infobloxopen/atlas-stubs/master/authn-stub/deploy/authn-stub.yaml | kubectl apply -f -
```

This will start AuthN stub that maps `User-And-Pass` header on JWT tokens, with following meaning:

```
admin1:admin -> AccountID=1
admin2:admin -> AccountID=2
```

##### Usage

Try it out by executing following curl commands when AuthN stub is running:

``` sh
# Create some profiles
curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/profiles -d '{"name": "personal", "notes": "Used for personal aims"}' | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/profiles -d '{"name": "work", "notes": "Used for work aims"}' | jq

# Create some groups assigned to profiles
curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/groups -d '{"name": "schoolmates", "profile_id": 1}' | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/groups -d '{"name": "family", "profile_id": 1}' | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/groups -d '{"name": "accountants", "profile_id": 2}' | jq

# Add some contacts assigned to profiles and groups
curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "primary_email": "mike@gmail.com", "profile_id": 1, "groups": [{"id": 1, "name": "schoolmates", "profile_id": 1}, {"id": 2, "name": "family", "profile_id": 1}], "home_address": {"city": "Minneapolis", "state": "Minnesota", "country": "US"}}' | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/contacts -d '{"first_name": "John", "primary_email": "john@gmail.com", "profile_id": 2, "work_address": {"city": "St.Paul", "state": "Minnesota", "country": "US"}}' | jq

# Read created resources
curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/profiles  | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/groups | jq

curl -k -H "User-And-Pass: admin1:admin" \
https://$(minikube ip)/atlas-contacts-app/v1/contacts | jq
```

The following commands without AuthN stub:
```bash
export JWT="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2NvdW50SUQiOjF9.GsXyFDDARjXe1t9DPo2LIBKHEal3O7t3vLI3edA7dGU"
curl -H "Authorization: Bearer $JWT" http://contacts.minikube/atlas-contacts-app/v1/contacts -d '{"first_name": "Mike", "primary_email": "mike@gmail.com"}'
curl -H "Authorization: Bearer $JWT" http://contacts.minikube/atlas-contacts-app/v1/contacts
```

##### Pagination (page token)

**DISCLAIMER**: it is intended only for demonstration purposes and should not be emulated.

Contacts App implements pagination in by adding application **specific** page token implementation.

Actually the service supports "composite" pagination in a specific way:

- limit and offset are still supported but without page token

- if an user requests page token and provides limit then limit value will be used as a step for all further requests
		`page_token = null & limit = 2 -> page_token=base64(offset=2:limit=2)`

- if an user requests page token and provides offset then only first time the provided offset is applied
		`page_token = null & offset = 2 & limit = 2 -> page_token=base64(offset=4:limit=2)`

Get all contacts: `GET http://localhost:8080/v1/contacts`
```json
{
  "results": [
    {
      "emails": [
        {
          "address": "one@mail.com",
          "id": "1"
        }
      ],
      "first_name": "Mike",
      "id": "1",
      "primary_email": "one@mail.com"
    },
    {
      "emails": [
        {
          "address": "two@mail.com",
          "id": "2"
        }
      ],
      "first_name": "Mike",
      "id": "2",
      "primary_email": "two@mail.com"
    },
    {
      "emails": [
        {
          "address": "three@mail.com",
          "id": "3"
        }
      ],
      "first_name": "Mike",
      "id": "3",
      "primary_email": "three@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK"
  }
}
```

Default pagination (supported by atlas-app-toolkit): `GET http://localhost:8080/v1/contacts?_limit=1&_offset=1`
```json
{
  "results": [
    {
      "emails": [
        {
          "address": "two@mail.com",
          "id": "2"
        }
      ],
      "first_name": "Mike",
      "id": "2",
      "primary_email": "two@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK"
  }
}
```

Request **specific** page token: `GET http://localhost:8080/v1/contacts?_page_token=null&_limit=2`
```json
{
  "results": [
    {
      "emails": [
        {
          "address": "one@mail.com",
          "id": "1"
        }
      ],
      "first_name": "Mike",
      "id": "1",
      "primary_email": "one@mail.com"
    },
    {
      "emails": [
        {
          "address": "two@mail.com",
          "id": "2"
        }
      ],
      "first_name": "Mike",
      "id": "2",
      "primary_email": "two@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK",
    "_page_token": "NDo0"
  }
}
```

Get next page via page token: `GET http://localhost:8080/v1/contacts?_page_token=NDo0`
```json
{
  "results": [
    {
      "emails": [
        {
          "address": "three@mail.com",
          "id": "3"
        }
      ],
      "first_name": "Mike",
      "id": "3",
      "primary_email": "three@mail.com"
    }
  ],
  "success": {
    "status": 200,
    "code": "OK",
    "_page_token": "NTo0"
  }
}
```

Get next page: `GET http://localhost:8080/v1/contacts?_page_token=NTo0`
The `"_page_token": "null"` means there are no more pages
```json
{
  "success": {
    "status": 200,
    "code": "OK",
    "_page_token": "null"
  }
}
```

## Deployment

Add additional notes about how to deploy this application. Maybe list some common pitfalls or debugging strategies.

## Running the tests

Explain how to run the automated tests for this system.

```
Give an example
```

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/your/project/tags).

## Traefik
A long time ago it seems that JohnB and I were working on the first
of this contacts app. JohnB had the AWS LB working with LoadBalancer.
We started to implement NGP API GW and decided to use NginX Ingress and
I worked on the first integration of this using a Beta version of NginX
Ingress, now it is a mature product. I see really that the NginX behavior
has not changed, after all NginX is trying to
[sell its Nginx+ product](https://www.loadbalancer.org/blog/nginx-vs-haproxy/).
So, it cripples its free opensource in a way that’s merely giving a taste
to small projects in the hope that once they grow
(together with their needs), they’ll stay hooked and buy into the system.

I started to debug a problem with my project that after hours could not
figure out :( I could have created a debug version of the Nginx Ingress
Controller to figure out in more detail why it was not working, but
instead use the chance to play with
[Traefik Ingress Controller](https://docs.traefik.io/user-guide/kubernetes/),
to see if it would be easier to deploy and debug. Traefik Ingress was
not available when we were doing our early work but is a more mature
product and unlike Nginx does not suffer from conflict of interest.

### Deployment
We start by removing Nginx from Minikube setup:
```bash
minikube addons disable ingress
minikube stop
minikube start
```
You can make sure Nginx is not running:
```bash
kubectl --namespace=kube-system get pods
```

Install Traefik by setting up the ClusterRoleBinding
```yaml
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
rules:
  - apiGroups:
      - ""
    resources:
      - services
      - endpoints
      - secrets
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - extensions
    resources:
      - ingresses
    verbs:
      - get
      - list
      - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: traefik-ingress-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: traefik-ingress-controller
subjects:
- kind: ServiceAccount
  name: traefik-ingress-controller
  namespace: kube-system
```
```bash
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-rbac.yaml
```

The user's manual has a good description of using Deployment or
DaemonSet, repeated below.

It is possible to use Træfik with a Deployment or a DaemonSet object,
whereas both options have their own pros and cons:

   * The scalability can be much better when using a Deployment,
   because you will have a Single-Pod-per-Node model when using a
   DaemonSet, whereas you may need less replicas based on your
   environment when using a Deployment.
   * DaemonSets automatically scale to new nodes, when the nodes
   join the cluster, whereas Deployment pods are only scheduled on
   new nodes if required.
   * DaemonSets ensure that only one replica of pods run on any
   single node. Deployments require affinity settings if you want
   to ensure that two pods don't end up on the same node.
   * DaemonSets can be run with the NET_BIND_SERVICE capability,
   which will allow it to bind to port 80/443/etc on each host.
   This will allow bypassing the kube-proxy, and reduce traffic hops.
   Note that this is against the Kubernetes Best Practices Guidelines,
   and raises the potential for scheduling/scaling issues.
   Despite potential issues, this remains the choice for most
   ingress controllers.
   * If you are unsure which to choose, start with the Daemonset.

I am using minikube and have only one node and will use a deployment
for the documentation below so you can follow along:
```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
---
kind: Deployment
apiVersion: extensions/v1beta1
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
  labels:
    k8s-app: traefik-ingress-lb
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: traefik-ingress-lb
  template:
    metadata:
      labels:
        k8s-app: traefik-ingress-lb
        name: traefik-ingress-lb
    spec:
      serviceAccountName: traefik-ingress-controller
      terminationGracePeriodSeconds: 60
      containers:
      - image: traefik
        name: traefik-ingress-lb
        ports:
        - name: http
          containerPort: 80
        - name: admin
          containerPort: 8080
        args:
        - --api
        - --kubernetes
        - --logLevel=INFO
---
kind: Service
apiVersion: v1
metadata:
  name: traefik-ingress-service
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
    - protocol: TCP
      port: 80
      name: web
    - protocol: TCP
      port: 8080
      name: admin
  type: NodePort
```
```bash
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-deployment.yaml
```
You should now see the Traefik Ingress Controller Pod running
```bash
kubectl --namespace=kube-system get pods
```
I created a kube-traefik.yaml to track the changes for this PoC,
the main change as you would guess is to the Ingress:
```yaml
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: contacts
  name: contacts-app-ingress
  annotations:
    kubernetes.io/ingress.class: traefik

spec:
  rules:
  - http:
      paths:
      - path: /atlas-contacts-app
        backend:
          serviceName: contacts-app
          servicePort: 8080
---
```
```bash
kubectl apply -f kube-traefik.yaml
```
Now I have an error when trying to use my contacts-app service.
It works from inside the cluster, but not from Ingress!
```bash
seizadi$ export JWT="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2NvdW50SUQiOjF9.GsXyFDDARjXe1t9DPo2LIBKHEal3O7t3vLI3edA7dGU"
seizadi$ curl -H "Authorization: Bearer $JWT" http://minikube/v1/contacts
curl: (7) Failed to connect to minikube port 80: Connection refused

seizadi$ k run -it --rm --image=infoblox/dnstools api-test
dnstools# export JWT="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2NvdW50SUQiOjF9.GsXyFDDARjXe1t9DPo2LIBKHEal3O7t3vLI3edA7dGU"
dnstools# curl -H "Authorization: Bearer $JWT" http://10.97.90.18:8080/v1/contacts | jq
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    38  100    38    0     0  12666      0 --:--:-- --:--:-- --:--:-- 12666
{
  "success": {
    "status": 200,
    "code": "OK"
  }
}
```
Lets start by creating a Service and an Ingress that will expose
the Træfik Web UI.
```yaml
apiVersion: v1
kind: Service
metadata:
  name: traefik-web-ui
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: traefik-web-ui
  namespace: kube-system
  annotations:
    kubernetes.io/ingress.class: traefik
spec:
  rules:
  - host: traefik-ui.minikube
    http:
      paths:
      - backend:
          serviceName: traefik-web-ui
          servicePort: 80


```
```bash
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/ui.yaml
```
Now lets setup an entry in our /etc/hosts file to route
traefik-ui.minikube to our cluster.

In production you would want to set up real DNS entries:
```bash
echo "$(minikube ip) traefik-ui.minikube" | sudo tee -a /etc/hosts
```
We should now be able to visit
[traefik-ui.minikube](http://traefik-ui.minikube/)
in the browser and view the Træfik web UI. I found that I could
not bring up the UI on the default port but I could on the
proxy port.
```bash
seizadi$ kubectl get services --namespace=kube-system
NAME                      CLUSTER-IP      EXTERNAL-IP   PORT(S)                       AGE
kube-dns                  10.96.0.10      <none>        53/UDP,53/TCP                 2d
kubernetes-dashboard      10.102.113.90   <nodes>       80:30000/TCP                  2d
traefik-ingress-service   10.105.72.46    <nodes>       80:30894/TCP,8080:31343/TCP   57m
```
So I was able to bring up UI on http://traefik-ui.minikube:31343,
so it seems to be a problem with Ingress Controller binding to NodePort
but not the default port.

I created a host for the contacts app:
```bash
echo "$(minikube ip) contacts.minikube" | sudo tee -a /etc/hosts
```
Now when I target the Nodeports I get a 404 error but not can not
get to service :(
```bash
seizadi$ curl -H "Authorization: Bearer $JWT" http://contacts.minikube:30894/v1/contacts
404 page not found
seizadi$ curl -H "Authorization: Bearer $JWT" http://contacts.minikube:31343/v1/contacts
404 page not found
```

So I removed the Traefik Deployment and created DaemonSet:
```yaml
The DaemonSet objects looks not much different:


---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
---
kind: DaemonSet
apiVersion: extensions/v1beta1
metadata:
  name: traefik-ingress-controller
  namespace: kube-system
  labels:
    k8s-app: traefik-ingress-lb
spec:
  template:
    metadata:
      labels:
        k8s-app: traefik-ingress-lb
        name: traefik-ingress-lb
    spec:
      serviceAccountName: traefik-ingress-controller
      terminationGracePeriodSeconds: 60
      containers:
      - image: traefik
        name: traefik-ingress-lb
        ports:
        - name: http
          containerPort: 80
          hostPort: 80
        - name: admin
          containerPort: 8080
        securityContext:
          capabilities:
            drop:
            - ALL
            add:
            - NET_BIND_SERVICE
        args:
        - --api
        - --kubernetes
        - --logLevel=INFO
---
kind: Service
apiVersion: v1
metadata:
  name: traefik-ingress-service
  namespace: kube-system
spec:
  selector:
    k8s-app: traefik-ingress-lb
  ports:
    - protocol: TCP
      port: 80
      name: web
    - protocol: TCP
      port: 8080
      name: admin
```
```bash
kubectl -n kube-system delete deployment traefik-ingress-controller
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-ds.yaml
serviceaccount "traefik-ingress-controller" configured
daemonset "traefik-ingress-controller" created
The Service "traefik-ingress-service" is invalid:
* spec.ports[0].nodePort: Forbidden: may not be used when `type` is 'ClusterIP'
* spec.ports[1].nodePort: Forbidden: may not be used when `type` is 'ClusterIP'
```
Now I can log into the UI properly but have the flagged errors with ClusterIP!!
To debug the above problem tried to use replace instead of apply:
```bash
kubectl replace -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-ds.yaml
serviceaccount "traefik-ingress-controller" replaced
Error from server (NotFound): error when replacing "https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-ds.yaml": daemonsets.extensions "traefik-ingress-controller" not found
Error from server (Invalid): error when replacing "https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-ds.yaml": Service "traefik-ingress-service" is invalid: spec.clusterIP: Invalid value: "": field is immutable
```
Looks like I removed the deployment but not the service,
so I removed the service and now DaemonSet works as expected:
```bash
k -n kube-system delete svc traefik-ingress-service
kubectl apply -f https://raw.githubusercontent.com/containous/traefik/master/examples/k8s/traefik-ds.yaml
Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply
serviceaccount "traefik-ingress-controller" configured
daemonset "traefik-ingress-controller" created
service "traefik-ingress-service" created
```
Now I can get back to the problem with contacts-app it ends up it was a
path problem, so I setup the Ingress like so:
```yaml
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: contacts
  name: contacts-app-ingress
  annotations:

spec:
  rules:
  - host: contacts.minikube
    http:
      paths:
#      - path: /atlas-contacts-app
      - path: /
        backend:
          serviceName: contacts-app
          servicePort: 8080
```
Now the curl works!
```bash
seizadi$ curl -H "Authorization: Bearer $JWT" http://contacts.minikube/v1/contacts
{"success":{"status":200,"code":"OK"}}
```
This is not acceptable we want the application to work with the
'/atlas-contacts-app' path and need more configuration for Ingress.
This is what the service annontation looks like to get this to work:
```yaml
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  namespace: contacts
  name: contacts-app-ingress
  annotations:
    kubernetes.io/ingress.class: traefik
    traefik.frontend.rule.type: PathPrefixStrip

spec:
  rules:
  - host: contacts.minikube
    http:
      paths:
      - path: /atlas-contacts-app
        backend:
          serviceName: contacts-app
          servicePort: 8080
```
The next step is to look at the AuthN stub that we have for NginX, it
looks like Traefik calls this feature
[Forward Authentication](https://docs.traefik.io/configuration/entrypoints/#forward-authentication)
need to figure out proper annotations to turn this on for
Traefik Ingress. Here are
[Ingress Forward Auth](https://stackoverflow.com/questions/50964605/traefik-forward-authentication-in-k8s-ingress-controller)
Looks like the feature is not in the latest 1.6 but in 1.7RC
[Traefik 1.7RC Doc](https://docs.traefik.io/v1.7/configuration/backends/kubernetes/#authentication)
We are running 1.6.5 which is the latest:
```bash
$ k -n kube-system logs traefik-ingress-controller-rqv65
time="2018-08-11T00:25:19Z" level=info msg="Traefik version v1.6.5 built on 2018-07-10_03:54:03PM"
```
