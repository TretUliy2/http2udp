pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        build 'build app'
      }
    }

  }
  environment {
    DB = 'mysql'
  }
}