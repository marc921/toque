resource "scaleway_vpc" "vpc" {
  name = "toque-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
	name = "toque-pn"
	vpc_id = scaleway_vpc.vpc.id
}