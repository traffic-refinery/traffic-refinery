# Traffic Refinery Overview

Traffic Refinery is a cost-aware network traffic analysis library implemented in Go

For a project overview, installation information, and detailed usage information please visit [Traffic Refinery's project homepage](https://traffic-refinery.github.io/)

# Installation

## Supported Operating Systems

* Debian Linux
* macOS
* The library should work in Windows, too, but has not being tested

## Dependencies

* [libpcap](https://www.tcpdump.org/) - Packet sniffing
* `[Optionally]` [PF_RING](https://www.ntop.org/products/packet-capture/pf_ring/)
* `[Optionally]` [AF_PACKET](http://manpages.org/af_packet/7)

## Citing Traffic Refinery

```
@article{10.1145/3491052,
    author = {Bronzino, Francesco and Schmitt, Paul and Ayoubi, Sara and Kim, Hyojoon and Teixeira, Renata and Feamster, Nick},
    title = {Traffic Refinery: Cost-Aware Data Representation for Machine Learning on Network Traffic},
    year = {2021},
    issue_date = {December 2021},
    publisher = {Association for Computing Machinery},
    address = {New York, NY, USA},
    volume = {5},
    number = {3},
    url = {https://doi.org/10.1145/3491052},
    doi = {10.1145/3491052},
    journal = {Proc. ACM Meas. Anal. Comput. Syst.},
    month = {dec},
    articleno = {40},
    numpages = {24}
}
```