#!/usr/bin/env groovy

// Automated release, promotion and dependencies
properties([
  // Include the automated release parameters for the build
  release.addParams(),
  // Dependencies of the project that should trigger builds
  dependencies([
    'cyberark/conjur-api-go',
  ])
])

// Performs release promotion.  No other stages will be run
if (params.MODE == "PROMOTE") {
  release.promote(params.VERSION_TO_PROMOTE) { sourceVersion, targetVersion, assetDirectory ->
    // Any assets from sourceVersion Github release are available in assetDirectory
    // Any version number updates from sourceVersion to targetVersion occur here
    // Any publishing of targetVersion artifacts occur here
    // Anything added to assetDirectory will be attached to the Github Release

    // Promote source version to target version.

    // NOTE: the use of --pull to ensure source images are pulled from internal registry
    sh "source ./bin/build_utils && ./bin/publish_container_images --promote --source ${sourceVersion}-\$(git_commit) --target ${targetVersion} --pull"
  }
  return
}

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
    timeout(time: 30, unit: 'MINUTES')
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  triggers {
    cron(getDailyCronString())
  }

  stages {
    // Aborts any builds triggered by another project that wouldn't include any changes
    stage ("Skip build if triggering job didn't create a release") {
      when {
        expression {
          MODE == "SKIP"
        }
      }
      steps {
        script {
          currentBuild.result = 'ABORTED'
          error("Aborting build because this build was triggered from upstream, but no release was built")
        }
      }
    }

    stage('Validate') {
      parallel {
        stage('Changelog') {
          steps { sh './bin/parse-changelog' }
        }
      }
    }

    // Generates a VERSION file based on the current build number and latest version in CHANGELOG.md
    stage('Validate changelog and set version') {
      steps {
        updateVersion("CHANGELOG.md", "${BUILD_NUMBER}")
      }
    }

    stage('Get latest upstream dependencies') {
      steps {
        updateGoDependencies("${WORKSPACE}/go.mod")
      }
    }

    stage('Build while unit testing') {
      parallel {
        stage('Run unit tests') {
          steps {
            sh './bin/test_unit'
          }
          post {
            always {
              sh './bin/coverage'
              junit 'junit.xml'

              cobertura autoUpdateHealth: false,
                autoUpdateStability: false,
                coberturaReportFile: 'coverage.xml',
                conditionalCoverageTargets: '70, 0, 0',
                failUnhealthy: false,
                failUnstable: false,
                maxNumberOfBuilds: 0,
                lineCoverageTargets: '70, 0, 0',
                methodCoverageTargets: '70, 0, 0',
                onlyStable: false,
                sourceEncoding: 'ASCII',
                zoomCoverageChart: false
                ccCoverage("gocov", "--prefix github.com/cyberark/conjur-cli-go")
            }
          }
        }

        stage('Build release artifacts') {
          steps {
            dir('./pristine-checkout') {
              // Go releaser requires a pristine checkout
              checkout scm

              // Copy VERSION info into prisitine folder
              sh "cp ../VERSION ./VERSION"

              // Create release artifacts without releasing to Github
              sh "./bin/build_release --skip-validate --clean"

              // Build container images
              sh "./bin/build_container_images"

              // Archive release artifacts
              archiveArtifacts 'dist/goreleaser/'
            }
          }
        }
      }
    }

    stage('Integration test while scanning') {
      parallel {
        stage('Run integration tests') {
          steps {
            dir('ci') {
              script {
                try{
                  sh 'summon -f ./okta/secrets.yml -e ci ./test_integration'
                } finally {
                  archiveArtifacts 'cleanup.log'
                }
              }
            }
          }
        }

         stage("Scan container images for fixable issues") {
           steps {
             scanAndReport("${containerImageWithTag()}", "HIGH", false)
           }
         }

        stage("Scan container images for total issues") {
          steps {
            scanAndReport("${containerImageWithTag()}", "NONE", true)
          }
        }
      }
    }

    stage('Release') {
      when {
        expression {
          MODE == "RELEASE"
        }
      }
      steps {
        release { billOfMaterialsDirectory, assetDirectory, toolsDirectory ->
          // Publish release artifacts to all the appropriate locations
          // Copy any artifacts to assetDirectory to attach them to the Github release

          // Copy assets to be published in Github release.
          sh "./bin/copy_release_artifacts ${assetDirectory}"

          // Create Go application SBOM using the go.mod version for the golang container image
          sh """go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --main "cmd/conjur/" --output "${billOfMaterialsDirectory}/go-app-bom.json" """
          // Create Go module SBOM
          sh """go-bom --tools "${toolsDirectory}" --go-mod ./go.mod --image "golang" --output "${billOfMaterialsDirectory}/go-mod-bom.json" """

          // Publish container images to internal registry
          sh './bin/publish_container_images --internal'
        }
      }
    }
  }
}

def containerImageWithTag() {
  sh(
    returnStdout: true,
    script: 'source ./bin/build_utils && echo "conjur-cli:$(project_version_with_commit)"'
  )
}
