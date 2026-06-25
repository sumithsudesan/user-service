variable "cluster_name" {
  type = string
}

variable "cluster_version" {
  type    = string
  default = "1.31"
}

variable "subnet_ids" {
  type = list(string)
}

variable "node_role_arn" {
  type = string
}

variable "node_instance_types" {
  type = list(string)
}

variable "node_desired" {
  type = number
}

variable "node_min" {
  type = number
}

variable "node_max" {
  type = number
}

variable "tags" {
  type    = map(string)
  default = {}
}
