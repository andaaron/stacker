load helpers

function setup() {
    cleanup
    rm -rf ocibuilds || true
    rm -rf oci_save || true
    mkdir -p ocibuilds/sub1
    mkdir oci_save
    touch ocibuilds/sub1/import1
    cat > ocibuilds/sub1/stacker.yaml <<EOF
layer1:
    from:
        type: docker
        url: docker://centos:latest
    import:
        - import1
    run: |
        cp /stacker/import1 /root/import1
EOF
    mkdir -p ocibuilds/sub2
    touch ocibuilds/sub2/import2
    cat > ocibuilds/sub2/stacker.yaml <<EOF
stacker_config:
    prerequisites:
        - ../sub1/stacker.yaml
    save_url: oci:oci_save
layer2:
    from:
        type: built
        tag: layer1
    import:
        - import2
    run: |
        cp /stacker/import2 /root/import2
        cp /root/import1 /root/import1_copied
EOF
    mkdir -p ocibuilds/sub3
    cat > ocibuilds/sub3/stacker.yaml <<EOF
stacker_config:
    prerequisites:
        - ../sub2/stacker.yaml
    save_url: oci:oci_save
layer3:
    from:
        type: built
        tag: layer2
    run: |
        cp /root/import2 /root/import2_copied
EOF
}

function teardown() {
    cleanup
    rm -rf ocibuilds || true
    rm -rf oci_save || true
}

@test "build layers and save them" {
    stacker build -f ocibuilds/sub3/stacker.yaml --remote-save

    # Determine expected commit hash
    commit_hash=commit-$(git rev-parse --short HEAD)
    [[ ! -z $(git status --porcelain --untracked-files=no) ]] && commit_hash=${commit_hash}-dirty
    user=$(git config user.email | cut -d"@" -f1)
    [[ ! -z ${user} ]] && commit_hash=${user}-${commit_hash}
    echo ${commit_hash}

    # Unpack saved image and check content
    mkdir dest
    umoci unpack --image oci_save:layer2_${commit_hash} dest/layer2
    [ "$status" -eq 0 ]
    [ -f dest/layer2/rootfs/root/import2 ]
    [ -f dest/layer2/rootfs/root/import1_copied ]
    [ -f dest/layer2/rootfs/root/import1 ]
    umoci unpack --image oci_save:layer3_${commit_hash} dest/layer3
    [ "$status" -eq 0 ]
    [ -f dest/layer3/rootfs/root/import2_copied ]
    [ -f dest/layer3/rootfs/root/import2 ]
    [ -f dest/layer3/rootfs/root/import1_copied ]
    [ -f dest/layer3/rootfs/root/import1 ]
}