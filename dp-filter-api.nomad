job "dp-filter-api" {
  datacenters = ["eu-west-1"]
  region      = "eu"
  type        = "service"

  // Make sure that this API is only ran on the publishing nodes
  constraint {
    attribute = "${node.class}"
    value     = "publishing"
  }

  group "web" {
    count = {{WEB_TASK_COUNT}}

    task "dp-filter-api" {
      driver = "exec"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/ons-dp-deployments/dp-filter-api/latest.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

         args = [
                  "${NOMAD_TASK_DIR}/dp-filter-api",
                ]
      }

      service {
        name = "dp-filter-api"
        port = "http"
        tags = ["web"]
      }

      resources {
        cpu    = "{{WEB_RESOURCE_CPU}}"
        memory = "{{WEB_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-filter-api"]
      }
    }
  }

  group "publishing" {
    count = {{PUBLISHING_TASK_COUNT}}

    task "dp-filter-api" {
      driver = "exec"

      artifact {
        source = "s3::https://s3-eu-west-1.amazonaws.com/ons-dp-deployments/dp-filter-api/latest.tar.gz"
      }

      config {
        command = "${NOMAD_TASK_DIR}/start-task"

         args = [
                  "${NOMAD_TASK_DIR}/dp-filter-api",
                ]
      }

      service {
        name = "dp-filter-api"
        port = "http"
        tags = ["publishing"]
      }

      resources {
        cpu    = "{{PUBLISHING_RESOURCE_CPU}}"
        memory = "{{PUBLISHING_RESOURCE_MEM}}"

        network {
          port "http" {}
        }
      }

      template {
        source      = "${NOMAD_TASK_DIR}/vars-template"
        destination = "${NOMAD_TASK_DIR}/vars"
      }

      vault {
        policies = ["dp-filter-api"]
      }
    }
  }

}