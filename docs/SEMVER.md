# Backwards compatibility promise

This document is a work in progress. 
The backwards compatibility promise will start as of version 1.0.0.
Before this version new release can break backwards compatibility.

## Semver

Kubicorn follows [Semantic Versioning](http://semver.org/) when creating new releases. 
In short, this means that we will only break backwards comparability when going major releases (such as 1.0, 2.0 etc).
When releasing minor(1.1.0, 1.2.0, etc) or patch updates(1.0.1, 1.2.8, etc) will not break backwards compatibility.

## Patch notes

Every new release will be accompanied by a list of all bugs fixed and features added in that release. 

## Updating major releases

When kubicorn decides to release a new major release we will include a specific document that documents all backwards compatibility breaks.
In this document the break and the recommended way to upgrade your environment will be explained.  

## Experimental

Releases can contain experimental features. 
These experimental features are excluded from the backwards compatibility promise. 
These features will marked as such to prevent confusion. 
Experimental features will not invalidate the backwards compatibility promise of existing features.