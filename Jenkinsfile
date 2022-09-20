#!/usr/bin/env groovy

pipeline {
  agent { label 'executor-v2' }

  options {
    timestamps()
    buildDiscarder(logRotator(numToKeepStr: '30'))
    timeout(time: 2, unit: 'HOURS')
  }

  environment {
    // Sets the MODE to the specified or autocalculated value as appropriate
    MODE = release.canonicalizeMode()
  }

  stages {
    stage('Validate CHANGELOG') {
      steps {
        sh './bin/parse-changelog'
      }
    }

    stage('Run Unit Tests') {
      steps {
        sh './bin/test_unit'
      }
    }
  }
}