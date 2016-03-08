KDeploy
=======
[![Build Status](https://drone.io/github.com/flexiant/kdeploy/status.png)](https://drone.io/github.com/flexiant/kdeploy/latest)

KDeploy, a tool that let's you deploy kubewares in your kubernetes cluster.

What is a kubeware?
-------------------

A kubeware is a combination of regular kubernetes API objects represented as `yaml` files, and a `metadata` file that contains the description and attributes for the application or service.

- `yaml` files are actual kubernetes files templates, that use [moustache](https://mustache.github.io/).
- `metadata` file contains basic metadata about the kubeware, references to all files that comprise the kubeware, and a list of managed attributes.
- `attributes.json` file that contains the values for those attributes in `yaml` files.

I'll give it a try, how do I create a kubeware?
-----------------------------------------------

Edit an existing kubernetes example, [guestbook](https://github.com/kubernetes/kubernetes/tree/master/examples/guestbook) if you will, and think of all those values that you'd like to modify for each environment. Let's say that when you deploy to a local vagrant cluster, you are good with `ServiceType` set to  `ClusterIP`, and that when you deploy to an environment that supports load balancing (like [Flexiant Concerto](https://start.concerto.io)) you want the `ServiceType` set to `LoadBalancer`.

You might also want to  change  values for the initial number of replicas, image names and tags, labels, ports, IPs, ... we can extract those values to an attribute file, and use `kdeploy` to combine the attribute file with kubernetes API files.

### Create Kubeware template

Hands on, edit a replication controller API file, and search for the replicas tag
```
spec:
  # this replicas value is default
  # modify it according to your case
  replicas: 2
```

Now substitute the number of replicas with a moustache placeholder.
```
spec:
  # this replicas value is default
  # modify it according to your case
  replicas: {{ $number_of_replicas }}
```

Modify other parameters and save the kubernetes API file. This template should remain untouched unless you change your application architecture.

### Create Metadata

We also need a `metadata.yaml` that serves as index for our kubeware and allow us to set default attributes.

Let's start adding some authoring metadata attributes:
- `name` name of the Kubeware.
- `maintainer`name of maintainer company or person(s).
- `email` contact email.
- `description` Kubeware description.
- `version` version of Kubeware will be used when deploying upgrades.
- `source` URL where Kubeware is to be found. Currently `kdeploy` only supports github.
- `issues` URL for the issue tracker.

```
name: "Guestbook"
maintainer: "Flexiant Ltd."
email: "contact@flexiant.com"
description: "Installs/Configures Guestbook Example via KDeploy"
version: "0.0.1"
source: "https://github.com/flexiant/kubeware-guestbook"
issues: "https://github.com/flexiant/kubeware-guestbook/issues"
```

Next, reference all kubernetes files in this Kubeware. Add service under `svc` and resource controllers under `rc`:

```
svc:
  redis-master: "redis-master-service.yaml"
  redis-slave: "redis-slave-service.yaml"
  frontend: "frontend-service.yaml"

rc:
  redis-master: "redis-master-controller.yaml"
  redis-slave: "redis-slave-controller.yaml"
  frontend: "frontend-controller.yaml"
```

Finally, create an `attributes` section, and add those attributes that need description, defaults, or required.
- Not all attributes must appear in this section. If an attribute doesn't appear in this section, it won't have a default, be required, nor have a description
- If an attribute is required and is not informed, or hasn't a default, `kdeploy` will return an error.

```
attributes:
  svc:
    frontend:
      balancer:
        description: "Defines how we want to expose the Frontend Service"
        default: LoadBalancer
        required: true
      port:
        description: "Defines expose port for the Frontend Service"
        default: 80
        required: true
    redis-master:
      port:
        description: "Defines expose port for the redis master service"
        default: 6379
        required: true
...
```

### Create Attributes file

Attributes files are customized for each environment. Let's create an "poor man's" scenario, where we don't want many replicas running.

```
{
  "rc":{
    "redis-slave":{
      "number":1
    ...
    }
  }
}
```
On the other hand wealthy Scrooge has plenty of nodes and can create a true redis cluster
```
{
  "rc":{
    "redis-slave":{
      "number":15
    }
  }
}
```

How do I use kdeploy?
---------------------
Upload the kubeware to a folder in github (we only support github for now)

To deploy the application with a set of attributes use
```
kdeploy deploy --kubeware https://github.com/flexiant/kubeware-guestbook --attribute poorman.json --namespace poorman
```

That will run a full application in kubernetes, using your selection of attributes.
You can use `kubectl`to check that the pods state, but we found useful adding listing and deleting behavior to kdeploy, so you don't have to switch tool.

To list all kubewares
```
kdeploy list --all
```

To delete a kubeware
```
kdeploy delete --kubeware https://github.com/flexiant/kubeware-guestbook --namespace poorman
```

What else can I do with kdeploy
-------------------------------
kdeploy was born as an internal tool to save time deploying k8s applications at [Flexiant](http://www.flexiant.com). We are planning to bring in some new features to kdeploy, as long as we keep it simple and agile.

Contributors are welcome anytime.
