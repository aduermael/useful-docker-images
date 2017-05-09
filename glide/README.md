# Glide

A container with [Glide](https://github.com/Masterminds/glide) (Vendor Package Management for Golang). 

Default command: `glide up` (updates vendor packages, using `glide.yaml`)

Within project repository:

```shell
docker run -v `pwd`:/proj -w /proj --rm glide
```

 