KDeploy
=======
[![Build Status](https://drone.io/github.com/flexiant/kdeploy/status.png)](https://drone.io/github.com/flexiant/kdeploy/latest)

KDeploy, a tool that let's you deploy kubewares in your kubernetes cluster.

What is a kubeware?
-------------------

A kubeware is a combination of regular kubernetes API objects represented as `yaml` files, and attributes that allows you to customize your environment.

I'll give it a try, how do I create a kubeware?
-----------------------------------------------

Edit an existing kubernetes example, [guestbook](https://github.com/kubernetes/kubernetes/tree/master/examples/guestbook) if you will, and think of all those values that you'd like to modify for each environment, let's say that for your local vagrant cluster, you are good with a `ServiceType` `ClusterIP`, and that when you deploy in an environment that supports load balancing (like [Flexiant Concerto](https://start.concerto.io)) you want the `ServiceType`to be `LoadBalancer`.

Also initial number of replicas, images, image tags, labels, ports, IPs, ... we can extract those values to an attribute  file, customized for each scenario, and use `kdeploy` to combine the attribute file with kubernetes API files.

Hands on, edit a replication controller API file, and search for the replicas tag
```
spec:
  # this replicas value is default
  # modify it according to your case
  replicas: 2
```

Now substitute the number of replicas with a placeholder using [go template](https://golang.org/pkg/text/template/) syntax.
```
spec:
  # this replicas value is default
  # modify it according to your case
  replicas: {{ $number_of_replicas }}
```

At the top of this file, assign the variable `$number_of_replicas` to a path in the forthcoming attribute file
```
{{ $number_of_replicas    := attr "rc/redis-slave/number" }}
```

Modify other parameters and save the kubernetes API file. This template should remain untouched unless you change your application architecture.

We also need a `metadata.yaml` that will add a description of the kubeware, it's components, it's default values, etc.

**TODO add metadata instructions**

That's not the case for the `attributes.json` file. The poor man's attributes file has only 1 redis slave replica
```
{
  "rc":{
    "redis-slave":{
      "number":1
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
kdeploy deploy --kubeware https://github.com/flexiant/kdeploy/tree/master/examples --attribute poorman.json --namespace poorman
```

That will run a full application in kubernetes, using your selection of attributes.
You can use `kubectl`to check that the pods state, but we found useful adding listing and deleting behavior to kdeploy, so you don't have to switch tool.

To list all kubewares
```
kdeploy list --all
```

To delete a kubeware
```
kdeploy delete --kubeware https://github.com/flexiant/kdeploy/tree/master/examples --namespace poorman
```

What else can I do with kdeploy
-------------------------------
kdeploy was born as an internal tool to save time deploying k8s applications at [Flexiant](http://www.flexiant.com). We are planning to bring in some new features to kdeploy, as long as we keep it simple and agile.

Contributors are welcome anytime.
