podTemplate (cloud:'sw-gpu-cloudnative',
    containers: [
    containerTemplate(name: 'docker', image: 'docker:dind', ttyEnabled: true, privileged: true),
    containerTemplate(name: 'golang', image: 'golang:1.16.3', ttyEnabled: true)
  ]) {
    node(POD_LABEL) {
        def scmInfo

        stage('checkout') {
            scmInfo = checkout(scm)
        }

        stage('dependencies') {
            container('golang') {
                sh 'GO111MODULE=off go get -u github.com/client9/misspell/cmd/misspell'
                sh 'GO111MODULE=off go get -u github.com/gordonklaus/ineffassign'
                sh 'GO111MODULE=off go get -u golang.org/x/lint/golint'
            }
            container('docker') {
                sh 'apk add --no-cache make bash git'
            }
        }
        stage('check') {
            parallel (
                getGolangStages(["assert-fmt", "lint", "vet", "ineffassign", "misspell"])
            )
        }
        stage('test') {
            parallel (
                getGolangStages(["test"])
            )
        }

        def versionInfo
        stage('version') {
            container('docker') {
                versionInfo = getVersionInfo(scmInfo)
                println "versionInfo=${versionInfo}"
            }
        }

        def dist = 'ubuntu20.04'
        def arch = 'amd64'
        def stageLabel = "${dist}-${arch}"

        stage('build-one') {
            container('docker') {
                stage (stageLabel) {
                    sh "make ${dist}-${arch}"
                }
            }
        }

        stage('release') {
            container('docker') {
                stage (stageLabel) {

                    def component = 'main'
                    def repository = 'sw-gpu-cloudnative-debian-local/pool/main/'

                    def uploadSpec = """{
                                        "files":
                                        [  {
                                                "pattern": "./dist/${dist}/${arch}/*.deb",
                                                "target": "${repository}",
                                                "props": "deb.distribution=${dist};deb.component=${component};deb.architecture=${arch}"
                                            }
                                        ]
                                    }"""

                    sh "echo starting release with versionInfo=${versionInfo}"
                    if (versionInfo.isTag) {
                        // upload to artifactory repository
                        def server = Artifactory.server 'sw-gpu-artifactory'
                        server.upload spec: uploadSpec
                    } else {
                        sh "echo skipping release for non-tagged build"
                    }
                }
            }
        }
    }
}

def getGolangStages(def targets) {
    stages = [:]

    for (t in targets) {
        stages[t] = getLintClosure(t)
    }

    return stages
}

def getLintClosure(def target) {
    return {
        container('golang') {
            stage(target) {
                sh "make ${target}"
            }
        }
    }
}

// getVersionInfo returns a hash of version info
def getVersionInfo(def scmInfo) {
    def versionInfo = [
        isTag: isTag(scmInfo.GIT_BRANCH)
    ]

    scmInfo.each { k, v -> versionInfo[k] = v }
    return versionInfo
}

def isTag(def branch) {
    if (!branch.startsWith('v')) {
        return false
    }

    def version = shOutput('git describe --all --exact-match --always')
    return version == "tags/${branch}"
}

def shOuptut(def script) {
    return sh(script: script, returnStdout: true).trim()
}
