# Backwards compatibility promise

This document is a work in progress. 
The backwards compatibility promise will start as of version 1.0.0.
Before this version new release can break backwards compatibility.

## Semver

Kubicorn follows [Semantic Versioning](http://semver.org/) when creating new releases. 
In short, this means that we will only break backwards compatibility when doing major releases (such as 1.0, 2.0 etc).
When releasing minor (1.1.0, 1.2.0, etc) or patch (1.0.1, 1.2.8, etc) updates no backward compatibility will be broken. 
Its important to us that Kubicorn keeps working as expected when updating you release between patches and updates. 
If any backwards compatibility is broken in a release, other then a major release, this will be rectified in a new update.

## Patch notes

Every new release will be accompanied by a list of all bugs fixed and features added in that release. 
Every bug and feature will have a link to the github issue associated with the change. 

## Updating major releases

When kubicorn decides to release a new major version we will include documentation that explains all backwards compatibility breaks.
The documentation wil also provider recommended ways to upgrade your environment to the new version of Kubicorn. 
Major updates don't have backwards compatibility by default.
A major release can release without breaking backwards compatibility.

## Experimental

Releases can contain experimental features. 
These experimental features are excluded from the backwards compatibility promise. 
These features will marked as such to prevent confusion. 
Experimental features will not invalidate the backwards compatibility promise of existing features.