# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

from setuptools import setup
from setuptools.dist import Distribution


class BinaryDistribution(Distribution):
    def has_ext_modules(self):
        return True


setup(
    name="micron-parser",
    version="1.1.0",
    description="Micron markup parser and HTML renderer (libmicron bindings)",
    license="0BSD",
    packages=["micron"],
    package_data={"micron": ["libmicron.so", "libmicron.dylib", "libmicron.dll"]},
    include_package_data=True,
    python_requires=">=3.9",
    distclass=BinaryDistribution,
    zip_safe=False,
)
