## Exercise 4 on Cloud Computing

### Summer Semester 2025



Dear students, 

After batteling with Docker and Docker Compose for the last couple of weeks, it is
time we level-up how we run our application. For this, we are gonna use a couple
of VMs to create a cluster using your Cloud provider and Kubernetes.

> This exercise can be completed using one VM functioning as master and worker
> node. However, you can create multiple instances and using them one as master,
> while the other one as the worker. The instructions here given will work for 
> both scenarios.

> If you want to test locally, please use minikube. For minikube, you only need
> docker installed in your system, while the setup is simple. Deploying resources
> uses the same YAML files as those for a K8s cluster.

### The Challenge

As you seek to deploy resources on K8s, the process looks completely different
compared to Docker Compose, especially the process is more manual. To successfully
pass this assignment, you must do:

1. Create a K8s cluster (one node is minimum).
2. Create the following for each of your microservices:
    
    - A deployment
    - A pod (given by the deployment)
    - A service

3. Install NGINX Ingress Controller (the Load Balancer) using helm to route the traffic to your pods as
done with Docker Compose in Exercise 3.
4. Create an ingress for each service to route the incoming traffic considering the request method 
(optional as you can configure the NGINX Ingress Controller using ConfigMaps)

#### Requirements

* You must create all images for `x86_64` (i.e., `amd64`).
* You must use a DB as a deployment. You can wrap your current MongoDB image into
a deployment to be used by you in your cluster.
* Use your unique code-bases to handle each operation (GET, POST, PUT, DELETE, and
server-side rendering).
* The URI for MongoDB must be given via an **environment variable**.
* The traffic must be redirected via **NGINX** by checking the request's method.
* Do not modify the headers inserted by NGINX.
* The automated tests from Exercise 1 must pass.
* You must expose execute `kubectl proxy` to enable remote inspection of your K8s cluster by running (this command runs as a process, which you can kill afterwards when you dont need it by running `ps aux | grep 'kubectl proxy'` and then running `sudo kill <pid>` where `<pid>` is the corresponding Process ID):
```bash
kubectl proxy --address='0.0.0.0' --port=<port> --accept-hosts='^*$' &
```
* You must open the `<port>` in your firewall rules
* You must submit the K8s Proxy endpoint with the format `http://<endpoint>:<port>`

#### Tests

- 40% of the grade comes from:

  1. Using an insolated MongoDB instance, the submission server will download, configure, and
  run all of your images. (10 pts).
  2. Using the isolated instance, the submission server will perform tests to the 
  different available endpoints (45 pts).
  3. Using your VM, the submission server will perform tests with the given endpoint
  and check that you are running **NGINX** with proper configuration (45 pts). Please
  make sure the port in the given port matches the exposed one for NGINX.

- 60% of the grade comes from:
  1. Using the K8s Proxy Endpoint, the submission server will query certain resource 
  descriptions to seek for the resources specified in "The Challenge" section.

### Important Information

#### Kubernetes Installation

For this assignment, you must use capable VMs, i.e., you can use `n1-standard-1` in `us-region-1`. That instance has a ca. $25/month cost. 

> You can also put your machine as part of the cluster using `kubeadm join`. 

> Remember to remove or shutdown your older VM to save resources.

Before installing Kubernetes in your machine, you need to install your container runtime. For this assignment, we will proceed using `containerd` as K8s does not support Docker directly. Since you are using a **Debian** image, you will follow [these instructions](https://docs.docker.com/engine/install/debian/):

> Remember you must follow these procedures for each machine you want to use K8s with.

```bash
sudo apt-get update
sudo apt-get install software-properties-common apt-transport-https ca-certificates gpg
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
```
Before you install containerd, first you need to setup the network to function with containerd by executing the following:

```bash
# K8s does not support swap memory
sudo swapoff -a 
sudo tee /etc/modules-load.d/containerd.conf << EOF
overlay
br_netfilter
EOF

sudo tee /etc/sysctl.d/kubernetes.conf << EOF
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
net.ipv4.ip_forward = 1
EOF

sudo sysctl --system
```

Once you this process finishes, then you can install containerd via:

```bash
sudo apt-get install -y containerd.io containernetworking-plugins
```

Once containerd has been installed, you need to configure it via:

```bash
mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml
sudo sed -i 's/SystemdCgroup \= false/SystemdCgroup \= true/g' /etc/containerd/config.toml
sudo systemctl daemon-reload
sudo systemctl enable containerd
sudo systemctl restart containerd
```
Now that contianerd is configured, we can proceed to install Kubernetes. For this exercise, we will be using the version `1.29`. To do that, you will need to follow [these instructions](https://v1-29.docs.kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/#installing-kubeadm-kubelet-and-kubectl):

```bash
# If the directory `/etc/apt/keyrings` does not exist, it should be created before the curl command, read the note below.
# sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# This overwrites any existing configuration in /etc/apt/sources.list.d/kubernetes.list
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl

sudo systemctl enable --now kubelet
```

##### For the master node
Once the previous step has completed, we can move to initialize the cluster by running:

```bash
# Sets the master node and initializes the K8s API Server, Proxy and other components
sudo kubeadm init --pod-network-cidr=10.244.0.0/16 --cri-socket=unix:///var/run/containerd/containerd.sock
```

After the previous command succeds, we want to execute `kubectl` in user mode, hence we need to grant our user access to the `admin token` used by K8s to communicate by:

```bash
mkdir -p $HOME/.kube
sudo cp -f /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

After this steps, we need to setup the networking in your cluster, during this exercise we rely on Flannel via:

```bash
# This command does not require sudo due to the previous commands. If kubectl cannot communicate with the api-server,
# check that $HOME/.kube/config has the right access
kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml
```

> Since we are running this exercise without other nodes, you also need to allow running pods on the master node via:
> `kubectl taint nodes --all node-role.kubernetes.io/master-``

> If you want to use a multi-node setup, you can follow the installation procedure in the other node and later generate a token
> for them to join your cluster via: `sudo kubeadm token create --print-join-command`

##### For the worker nodes

Once the cluster has been created, you must follow the installations steps laid before, and with the generated token, you run it on each worker node as:

```bash
sudo kubeadm -v 5 join <ip_master>:6443 --token <token> --discovery-token-ca-cert-hash <hash>
```

#### Monitoring resources on K8s

Once you have initialized and configure your cluster, you can fetch the status of your resources: from nodes to pods, you can do that by executing

```bash
kubectl get <resource>
```

where `<resource>` could take the value of `nodes`, `pods`, `svc` (short for services), `deployments`. For all available options, you list them by running

```bash
kubectl get --help
```

In case your resources are namespaced, the previous command will only return those within the "current context", the default current context is set to the namespace `default`. If you want to observe resources from other namespaces, you can use the `kubectl` command as:

```bash
kubectl <op> -n <namespace> <resource>
```

where `<op>`can take the value of `get`, `edit`, `describe`, `delete`. For logs, you can use:

```bash
# You can append an -f to this command to follow the logs
kubectl logs -n <namespace> <pod-name>
```

Upon correct configuration of your cluster (and before deploying any more resources), executing `kubectl get pods -A` will return:

```txt
NAME                              READY   STATUS    RESTARTS      AGE
coredns-76f75df574-5cdsb          1/1     Running   0             106d
coredns-76f75df574-wrqmq          1/1     Running   0             106d
etcd-capsvm2                      1/1     Running   0             106d
kube-apiserver-capsvm2            1/1     Running   0             94d
kube-controller-manager-capsvm2   1/1     Running   5 (20h ago)   106d
kube-proxy-2q4pz                  1/1     Running   0             98d
kube-proxy-f6kzw                  1/1     Running   0             2d4h
kube-proxy-fpbpr                  1/1     Running   0             106d
kube-proxy-gzj56                  1/1     Running   0             47d
kube-proxy-hxtqn                  1/1     Running   0             106d
kube-proxy-ld99h                  1/1     Running   0             105d
kube-proxy-sc9rd                  1/1     Running   0             47d
kube-proxy-sv44k                  1/1     Running   0             93d
kube-proxy-t7tlt                  1/1     Running   0             106d
kube-proxy-zv9r2                  1/1     Running   0             47d
kube-scheduler-capsvm2            1/1     Running   6 (20h ago)   106d
```

> Do not worry about those restarts for kube-controller-manager and kube-scheduler.

For more information on how you can use `kubectl`, please read [this guide](https://kubernetes.io/docs/reference/kubectl/) by Kubernetes.

#### Installing Helm

Before we install NGINX Ingress Controller, we need to install Helm using [these instructions](https://helm.sh/docs/intro/install/):

```bash
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
```

#### Installing NGINX Ingress Controller

Once Helm has been installed, we can install NGINX by following [these instructions](https://docs.nginx.com/nginx-ingress-controller/installation/installing-nic/installation-with-helm/#installing-the-chart):

```bash
# Please change my-release to nginx-<desired-suffix>
helm install my-release oci://ghcr.io/nginxinc/charts/nginx-ingress --version 1.2.2
```

> The following section requires that you have deployed your apps before hand

However, since you want to do re-routing based on the HTTP method, you might have to change the `ConfigMap` for the NGINX Ingress Controller. You can do so by executing the following commands:

```bash
# Returns all ConfigMaps across namespaces
kubectl get cm -A

# Edit the ConfigMap for NGINX
kubectl edit cm -n <namespace> <cm-name>
```

You need to specify one of values given in [the NGINX Ingress Controller documentation](https://docs.nginx.com/nginx-ingress-controller/configuration/global-configuration/configmap-resource/#snippets-and-custom-templates). For example:

```yaml
http-snippets: |
  upstream my-service {
    server my-service.<namespace>.svc.cluster.local:<port>;
  }
  ...
```

> However, a easier way to use the `NGINX Ingress Controller` is by specifying a [main-template](https://github.com/nginx/kubernetes-ingress/tree/v5.0.0/examples/shared-examples/custom-templates). 
> It basically replaces the default `nginx.conf` file. This is the simples method. It also gives you the freedom to
> setup the controller as you used it in previous assignments.

Another way is to create an Ingress for your services and rewrite the path based on the incoming traffic. For example, using [this example](https://kubernetes.github.io/ingress-nginx/user-guide/basic-usage/) from NGINX:

> Remember you need to save this content in a file and then apply it via `kubectl apply -f <nginx-ingress>`

```yaml
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress-myserviceb
  annotations:
    # use the shared ingress-nginx
    kubernetes.io/ingress.class: "nginx"
spec:
  rules:
  - host: myserviceb.foo.org
    http:
      paths:
      - path: /
        backend:
          serviceName: myserviceb
          servicePort: 80
```

You can adapt it using the helpful comments from [this Github comment](https://github.com/kubernetes/ingress-nginx/issues/187#issuecomment-576708564), which means you only need one ingress for all of your services, and that ingress performs the routing.

```yaml
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/configuration-snippet: |
      internal;
      rewrite ^ $original_uri break;
    nginx.ingress.kubernetes.io/server-snippet: |
      location /api/does/not/matter/much {
        if ( $request_method = GET) {
          set $target_destination '/_read';
        }
        if ( $request_method != GET) {
          set $target_destination '/_write';
        }
        set $original_uri $uri;
        rewrite ^ $target_destination last;
      }
...
paths:
  - path: /_read
    backend:
      serviceName: read-service
      servicePort: 8080
  - path: /_write
    backend:
      serviceName: write-service
      servicePort: 8080
```

Once you have deploy the NGINX Ingress Controller, you need to manually assign an external IP to the Load Balancer. For this task, you will have to execute the following:

```bash
# This command will show the services for the "default" namespace
kubectl get svc
```

From the output, you will copy the name for the NGINX Ingress Controller service and will perform the following:

```bash
kubectl edit svc <name>
```

Running this command will open vim with the YAML file for the service; it will look something like this:

```yaml
kind: Service
apiVersion: v1
metadata:
  name: ingress-nginx
  namespace: ingress-nginx
  labels:
    app: ingress-nginx
spec:
  selector:
    app: ingress-nginx
  ports:
  - name: http
    port: 80
    targetPort: http
  - name: https
    port: 443
    targetPort: http
  ...
```

You will need to add the section:

```yaml
externalIPs:
- <internalIP>
```

where `<internalIP>` is the internal IP shown in your Google Console. After the changes, it should look something like:

```yaml
kind: Service
apiVersion: v1
metadata:
  name: ingress-nginx
  namespace: ingress-nginx
  labels:
    app: ingress-nginx
spec:
  selector:
    app: ingress-nginx
  ports:
  - name: http
    port: 80
    targetPort: http
  - name: https
    port: 443
    targetPort: http
  externalIPs:
  - <internalIP>
  ...
```

Once you have made these changes, you can exit and save the changes by pressing `Ctrl+c`, then `:wqa`. You should see an output that the service has been modified.

After the effect, you can run `kubectl get svc` once again, and the NGINX Ingress Service will have a valid external IP.

#### Deploying your applications

To deploy your applications, you need to create a couple of files, specifically a YAML file for your `deployment` and other for your `service` per pod you want to create. The template for a deployment is the following:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: <deployment-unique-name>
spec:
  replicas: 1
  selector:
    matchLabels:
      app: <deployment-unique-name>
  template:
    metadata:
      labels:
        app: <deployment-unique-name>
    spec:
      containers:
        - name: <deployment-unique-name>
          image: <image>
          # You can also use IfNotPresent but it might lead to old images
          imagePullPolicy: "Always"
          env:
            - name: TZ
              value: Europe/Berlin
          ports:
            - containerPort: <internalPort>
```

```yaml
apiVersion: v1
kind: Service
metadata:
  name: <deployment-unique-name>-service
spec:
  type: ClusterIP
  ports:
    - name: <deployment-unique-name>-endpoint
      port: <externalPort>
      targetPort: <internalPort>
      protocol: TCP
  selector:
    app: <deployment-unique-name>
```

> You need to submit the `<deployment-unique-name>` to the submission portal for each deplyoment you create (GET, POST, PUT, DELETE, and rendering)

Once you have created the corresponding `deployment` and `service` file per µservice, you can create the resources by running `kubectl apply -f <µservice-deployment>` per µservice, and later on `kubectl apply -f <µservice-service>` per µservice.

Upon deploying your resources, NGINX Ingress Controller, and Ingresses; you can check the status by executing 

```bash
kubectl get pods -A
```

and all pods should show `Running` as their status. If they are not like that, please reach me via Piazza.

#### Happy Coding!
