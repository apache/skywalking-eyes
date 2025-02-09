# Apache SkyWalking Eyes Release Guide

This documentation guides the release manager to release the SkyWalking Eyes in the Apache Way, and also helps people to check the release for voting.

## Prerequisites

1. Close (if finished, or move to next milestone otherwise) all issues in the current milestone from [skywalking-eyes](https://github.com/apache/skywalking-eyes/milestones) and [skywalking](https://github.com/apache/skywalking/milestones), create a new milestone if needed.

## Add your GPG public key to Apache svn

1. Upload your GPG public key to a public GPG site, such as [MIT's site](http://pgp.mit.edu:11371/). 

1. Log in [id.apache.org](https://id.apache.org/) and submit your key fingerprint.

1. Add your GPG public key into [SkyWalking GPG KEYS](https://dist.apache.org/repos/dist/release/skywalking/KEYS) file, **you can do this only if you are a PMC member**.  You can ask a PMC member for help. **DO NOT override the existed `KEYS` file content, only append your key at the end of the file.**

## Build and sign the source code package

```shell
export VERSION=<the version to release>

git clone --recurse-submodules git@github.com:apache/skywalking-eyes && cd skywalking-eyes
git tag -a "v$VERSION" -m "Release Apache SkyWalking-Eyes $VERSION"
git push --tags

make clean 
make verify # this is optional, it runs sanity checks to verify the features
make release

# Create a draft release in GitHub, this can be optionally done in GitHub web ui.
gh release create v$VERSION --generate-notes --draft skywalking-license-eye-$VERSION-*
```

## Upload to Apache svn

```bash
svn co https://dist.apache.org/repos/dist/dev/skywalking/
mkdir -p skywalking/eyes/"$VERSION"
cp skywalking-eyes/skywalking*.tgz skywalking/eyes/"$VERSION" 
cp skywalking-eyes/skywalking*.tgz.asc skywalking/eyes/"$VERSION"
cp skywalking-eyes/skywalking-license-eye*.tgz.sha512 skywalking/eyes/"$VERSION"

cd release/skywalking && svn add eyes/$VERSION && svn commit eyes -m "Draft Apache SkyWalking-Eyes release $VERSION"
```

## Call for vote in dev@ mailing list

Call for vote in `dev@skywalking.apache.org`.

```text
Subject: [VOTE] Release Apache SkyWalking Eyes version $VERSION

Content:

Hi the SkyWalking Community,

This is a call for vote to release Apache SkyWalking Eyes version $VERSION.

Release notes:

 * https://github.com/apache/skywalking-eyes/blob/v$VERSION/CHANGES.md

Release Candidate:

 * https://dist.apache.org/repos/dist/dev/skywalking/eyes/$VERSION
 * sha512 checksums
   - sha512xxxxyyyzzz skywalking-eyes-src-x.x.x.tgz

Release Tag :

 * (Git Tag) v$VERSION

Release Commit Hash :

 * https://github.com/apache/skywalking-eyes/tree/<Git Commit Hash>

Keys to verify the Release Candidate :

 * https://dist.apache.org/repos/dist/release/skywalking/KEYS

Guide to build the release from source :

 * https://github.com/apache/skywalking-eyes/blob/v$VERSION/docs/How-to-release.md#build-and-sign-the-source-code-package

Voting will start now and will remain open for at least 72 hours, all PMC members are required to give their votes.

[ ] +1 Release this package.
[ ] +0 No opinion.
[ ] -1 Do not release this package because....

Thanks.

[1] https://github.com/apache/skywalking-eyes/blob/main/docs/How-to-release.md#vote-check
```

## Vote Check

All PMC members and committers should check these before voting +1:

1. Features test.
1. All artifacts in staging repository are published with `.asc`, `.md5`, and `sha` files.
1. Source codes and distribution packages (`skywalking-eyes-src-$VERSION.tgz`)
are in `https://dist.apache.org/repos/dist/dev/skywalking/eyes/$VERSION` with `.asc`, `.sha512`.
1. `LICENSE` and `NOTICE` are in source codes and distribution package.
1. Check `shasum -c skywalking-eyes-src-$VERSION.tgz.sha512`.
1. Check `gpg --verify skywalking-eyes-src-$VERSION.tgz.asc skywalking-eyes-src-$VERSION.tgz`.
1. Build distribution from source code package by following this [the build guide](#build-and-sign-the-source-code-package).
1. Licenses check, `make license`.

Vote result should follow these:

1. PMC vote is +1 binding, all others is +1 no binding.

1. Within 72 hours, you get at least 3 (+1 binding), and have more +1 than -1. Vote pass. 

1. **Send the closing vote mail to announce the result**.  When count the binding and no binding votes, please list the names of voters. An example like this:

   ```
   [RESULT][VOTE] Release Apache SkyWalking Eyes version $VERSION
   
   72+ hours passed, we’ve got ($NUMBER) +1 bindings (and ... +1 non-bindings):
   
   (list names)
   +1 bindings:
   xxx
   ...
   
   +1 non-bindings:
   xxx
   ...
    
   Thank you for voting, I’ll continue the release process.
   ```

## Publish release

1. Move source codes tar balls and distributions to `https://dist.apache.org/repos/dist/release/skywalking/`, **you can do this only if you are a PMC member**.

    ```shell
    svn mv https://dist.apache.org/repos/dist/dev/skywalking/eyes/"$VERSION" https://dist.apache.org/repos/dist/release/skywalking/eyes/"$VERSION"
    ```
   
2. Push Docker images.

    ```shell
    export VERSION=<the version to release> 
    export PROJECT=skywalking-eyes
    docker login
    make docker-release
    ```
    
3. Refer to the previous [PR](https://github.com/apache/skywalking-website/pull/132), update news and links on the website. There are several files need to modify.

4. Update [Github release page](https://github.com/apache/skywalking-eyes/releases), follow the previous convention.

5. Send ANNOUNCE email to `dev@skywalking.apache.org` and `announce@apache.org`, the sender should use his/her Apache email account. 

    ```
    Subject: [ANNOUNCEMENT] Apache SkyWalking Eyes $VERSION Released

    Content:

    Hi the SkyWalking Community,

    On behalf of the SkyWalking Team, I’m glad to announce that SkyWalking Eyes $VERSION is now released.

    SkyWalking Eyes: A full-featured license tool to check and fix license headers and resolve dependencies' licenses.

    SkyWalking: APM (application performance monitor) tool for distributed systems, especially designed for microservices, cloud native and container-based (Docker, Kubernetes, Mesos) architectures.

    Download Links: https://skywalking.apache.org/downloads/

    Release Notes : https://github.com/apache/skywalking-eyes/blob/v$VERSION/CHANGES.md

    Website: https://skywalking.apache.org/

    SkyWalking Eyes Resources:
    - Issue: https://github.com/apache/skywalking/issues
    - Mailing list: dev@skywalking.apache.org
    - Documents: https://github.com/apache/skywalking-eyes/blob/v$VERSION/README.md
    
    The Apache SkyWalking Team
   ```
    
    
