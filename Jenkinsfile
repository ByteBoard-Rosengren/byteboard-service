pipeline {
      agent any

      environment {
          DOCKER_IMAGE = 'krosengr4/byteboard-service'
          DOCKER_CREDENTIALS_ID = 'dockerhub-creds'
      }

      stages {
          stage('Checkout') {
              steps {
                  checkout scm
              }
          }

          stage('Build Docker Image') {
              steps {
                  script {
                      echo 'Building Docker image...'
                      sh "docker build -t ${DOCKER_IMAGE}:${BUILD_NUMBER} -t ${DOCKER_IMAGE}:latest ."
                  }
              }
          }

          stage('Push to Docker Hub') {
              steps {
                  script {
                      echo 'Pushing to Docker Hub...'
                      withCredentials([usernamePassword(
                          credentialsId: DOCKER_CREDENTIALS_ID,
                          usernameVariable: 'DOCKER_USER',
                          passwordVariable: 'DOCKER_PASS'
                      )]) {
                          sh '''
                              echo $DOCKER_PASS | docker login -u $DOCKER_USER --password-stdin
                              docker push ${DOCKER_IMAGE}:${BUILD_NUMBER}
                              docker push ${DOCKER_IMAGE}:latest
                          '''
                      }
                  }
              }
          }
      }

      post {
          success {
              echo "Build successful! Image: ${DOCKER_IMAGE}:${BUILD_NUMBER}"
              build job: 'RosenPi/Byteboard-Service/deploy', wait: false
          }
          failure {
              echo 'Build failed!'
          }
          always {
              sh 'docker logout'
          }
      }
  }
