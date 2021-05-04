# Release workflow

1. Create the release branch
2. Update the project version and tag
3. Build the artifacts
4. Verify the artifacts
5. Move master branch to the next version
6. Write release notes

# Steps in detail

1. Create the release branch

```
git clone https://github.com/streamnative/function-mesh
git checkout -b branch-x.y
cd mesh-worker-service
mvn versions:set -DnewVersion=vx.y.z
git add .
git commit -m "Update release version for mesh worker service"
git push origin branch-x.y
```

2. Update the project version and tag

```
git tag vX.Y.Z
git push origin vX.Y.Z
```

3. Click the release button

Click the release button and draft a new release. When publish the release, the Action CI will automatically trigger the release process, build the corresponding image, and push it to docker_hub.
