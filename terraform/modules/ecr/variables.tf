variable "repositories" {
  type = list(string)
}

variable "image_retention" {
  type    = number
  default = 10
}

variable "tags" {
  type    = map(string)
  default = {}
}
