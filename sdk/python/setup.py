#!/usr/bin/env python3
"""
FlyPy Python SDK - Deterministic Python Compilation

A Python SDK for compiling deterministic Python functions to WebAssembly
for execution on the FunctionFly platform.
"""

from setuptools import setup, find_packages
import os

# Read version from version file
def read_version():
    version_file = os.path.join(os.path.dirname(__file__), 'flypy', 'version.py')
    with open(version_file, 'r') as f:
        exec(f.read())
        return locals()['__version__']

def read_readme():
    with open('README.md', 'r', encoding='utf-8') as f:
        return f.read()

setup(
    name="flypy",
    version=read_version(),
    author="FunctionFly Team",
    author_email="team@functionfly.com",
    description="Deterministic Python compilation to WebAssembly",
    long_description=read_readme(),
    long_description_content_type="text/markdown",
    url="https://github.com/functionfly/functionfly",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 3 - Alpha",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
        "Topic :: Software Development :: Compilers",
        "Topic :: Software Development :: Libraries :: Python Modules",
    ],
    python_requires=">=3.8",
    install_requires=[
        "typing-extensions>=4.12.0",
        "pydantic>=2.9.0",
        "jsonschema>=4.23.0",
    ],
    extras_require={
        "dev": [
            "pytest>=8.0.0",
            "black>=24.0.0",
            "mypy>=1.11.0",
            "flake8>=7.0.0",
        ],
        "build": [
            "build>=1.0.0",
            "twine>=5.0.0",
        ],
    },
    entry_points={
        "console_scripts": [
            "flypy=flypy.cli:main",
        ],
    },
    include_package_data=True,
    zip_safe=False,
)