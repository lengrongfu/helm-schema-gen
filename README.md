# helm schema gen plugin [ CURRENTLY NOT MAINTAINED ]

![](https://github.com/lengrongfu/helm-schema-gen/workflows/goreleaser/badge.svg)

So that you don't have to write values.schema.json by hand from scratch for your Helm 3 charts

[Helm](https://helm.sh) plugin to
generate [JSON Schema for values yaml](https://helm.sh/docs/topics/charts/#schema-files)


#"重要"提示提示框，在"重要："后填上你想要的内容，例如"这是一个示例。"
<div style="padding: 15px; border: 1px solid transparent; border-color: transparent; margin-bottom: 20px; border-radius: 4px; color: #31708f; background-color: #d9edf7; border-color: #bce8f1;">
&#x1F50A<b> Important Notice </b>
</div>

> Because the [upstream project](https://github.com/karuppiah7890/helm-schema-gen) has been archived, subsequent updates will be submitted to the current repository.
> Important updates are recorded below:
> 1. Added json sorting output to keep the format order consistent with that in the yaml file.

## Note about maintenance

I currently don't have the bandwidth to reply to issues, write code and review PRs. For now I recommend forking the repo
and making changes and using the fork 😅

## Code stuff

Nothing fancy about the code, all the heavy lifting is done by:

- [go-jsonschema-generator](https://github.com/lengrongfu/go-jsonschema-generator) - for generating JSON schema. It's a
  fork of [this](https://github.com/mcuadros/go-jsonschema-generator). Thanks
  to [@mcuadros](https://github.com/mcuadros)
- [go-yaml](https://github.com/go-yaml/yaml/) - for YAML parsing
- [cobra](https://github.com/spf13/cobra) - for CLI stuff
- [The Go stdlib](https://golang.org/pkg/) - for everything else

## Install

The plugin works with both Helm v2 and v3 versions as it's agnostic to the Helm
binary version

```
$ helm plugin install https://github.com/lengrongfu/helm-schema-gen.git
lengrongfu/helm-schema-gen info checking GitHub for tag '0.0.4'
lengrongfu/helm-schema-gen info found version: 0.0.4 for 0.0.4/Darwin/x86_64
lengrongfu/helm-schema-gen info installed ./bin/helm-schema-gen
Installed plugin: schema-gen
```

But note that the schema feature is present only in Helm v3 charts, so Helm
chart still has to be v3, meaning - based on the Helm chart v3 spec. And the
schema validation is only done in Helm v3. Read more in the
[Schema Files](https://helm.sh/docs/topics/charts/#schema-files) section of the
Helm official docs.

## Usage

The plugin works with both Helm v2 and v3 versions

Let's take a sample `values.yaml` like the below

```
replicaCount: 1

image:
  repository: nginx
  pullPolicy: IfNotPresent

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
  # Specifies whether a service account should be created
  create: true
  # The name of the service account to use.
  # If not set and create is true, a name is generated using the fullname template
  name:

podSecurityContext: {}
  # fsGroup: 2000

securityContext: {}
  # capabilities:
  #   drop:
  #   - ALL
  # readOnlyRootFilesystem: true
  # runAsNonRoot: true
  # runAsUser: 1000

service:
  type: ClusterIP
  port: 80

ingress:
  enabled: false
  annotations: {}
    # kubernetes.io/ingress.class: nginx
    # kubernetes.io/tls-acme: "true"
  hosts:
    - host: chart-example.local
      paths: []
  tls: []
  #  - secretName: chart-example-tls
  #    hosts:
  #      - chart-example.local

resources: {}
  # We usually recommend not to specify default resources and to leave this as a conscious
  # choice for the user. This also increases chances charts run on environments with little
  # resources, such as Minikube. If you do want to specify resources, uncomment the following
  # lines, adjust them as necessary, and remove the curly braces after 'resources:'.
  # limits:
  #   cpu: 100m
  #   memory: 128Mi
  # requests:
  #   cpu: 100m
  #   memory: 128Mi

nodeSelector: {}

tolerations: []

affinity: {}
```

Now if you use the plugin and pass the `values.yaml` to it, you will
get the JSON Schema for the `values.yaml`

```
$ helm schema-gen values.yaml
{
  "$schema": "http://json-schema.org/schema#",
  "type": "object",
  "properties": {
    "replicaCount": {
      "type": "integer"
    },
    "image": {
      "type": "object",
      "properties": {
        "repository": {
          "type": "string"
        },
        "pullPolicy": {
          "type": "string"
        },
        "tag": {
          "type": "string"
        }
      }
    },
    "imagePullSecrets": {
      "type": "array"
    },
    "nameOverride": {
      "type": "string"
    },
    "fullnameOverride": {
      "type": "string"
    },
    "serviceAccount": {
      "type": "object",
      "properties": {
        "create": {
          "type": "boolean"
        },
        "annotations": {
          "type": "object"
        },
        "name": {
          "type": "string"
        }
      }
    },
    "podAnnotations": {
      "type": "object"
    },
    "podSecurityContext": {
      "type": "object"
    },
    "securityContext": {
      "type": "object"
    },
    "service": {
      "type": "object",
      "properties": {
        "type": {
          "type": "string"
        },
        "port": {
          "type": "integer"
        }
      }
    },
    "ingress": {
      "type": "object",
      "properties": {
        "hosts": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "host": {
                "type": "string"
              },
              "paths": {
                "type": "array"
              }
            }
          }
        },
        "tls": {
          "type": "array"
        },
        "enabled": {
          "type": "boolean"
        },
        "annotations": {
          "type": "object"
        }
      }
    },
    "resources": {
      "type": "object"
    },
    "autoscaling": {
      "type": "object",
      "properties": {
        "enabled": {
          "type": "boolean"
        },
        "minReplicas": {
          "type": "integer"
        },
        "maxReplicas": {
          "type": "integer"
        },
        "targetCPUUtilizationPercentage": {
          "type": "integer"
        }
      }
    },
    "nodeSelector": {
      "type": "object"
    },
    "tolerations": {
      "type": "array"
    },
    "affinity": {
      "type": "object"
    }
  }
}
```

You can save it to a file like this

```
$ helm schema-gen values.yaml > values.schema.json
```

## Issues? Feature Requests? Proposals? Feedback?

Note: I currently don't have the bandwidth to reply to issues, write code and review PRs. For now I recommend forking
the repo and making changes and using the fork 😅

Put them all in [GitHub issues](https://github.com/lengrongfu/helm-schema-gen/issues) 😁
I value every feedback. I really want to make sure that my tools help people and does not
annoy people. I want my tools to enable people and not hinder them. I'll do my best to help you
if you face any hindrance because of using my tools! :)
