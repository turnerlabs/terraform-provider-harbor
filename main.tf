provider "harbor" {
  credentials = "${file("~/.harbor/credentials")}"
}
