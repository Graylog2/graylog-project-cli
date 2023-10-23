if (currentBuild.buildCauses.toString().contains('BranchIndexingCause'))
{
  print "Build skipped due to trigger being Branch Indexing."
  return
}


pipeline
{
   agent { label 'linux' }

   options
   {
      buildDiscarder logRotator(artifactDaysToKeepStr: '30', artifactNumToKeepStr: '10', daysToKeepStr: '30', numToKeepStr: '10')
      timestamps()
   }

   tools
   {
     go 'Go 1.21'
   }

   stages
   {
      stage('Build')
      {
         steps {
            sh 'go version'

            sh "make"
         }
      }

      stage('Sign Windows Binaries')
      {
        agent
        {
          docker
          {
            image 'graylog/internal-codesigntool:latest'
            args '-u jenkins:jenkins'
            registryCredentialsId 'docker-hub'
            alwaysPull true
            reuseNode true
          }
        }

        environment
        {
          CODESIGN_USER = credentials('codesign-user')
          CODESIGN_PASS = credentials('codesign-pass')
          CODESIGN_TOTP_SECRET = credentials('codesign-totp-secret')
          CODESIGN_CREDENTIAL_ID = credentials('codesign-credential-id')
        }

        steps
        {
          sh 'make sign-binaries'
        }

      }

      stage('Release')
      {

         when
         {
             buildingTag()
         }

         environment
         {
             GITHUB_CREDS = credentials('github-access-token')
         }

         steps
         {
           echo "Releasing ${TAG_NAME} to Github..."

           script
           {
             def RELEASE_DATA = sh returnStdout: true, script: "curl -fsSL --user \"$GITHUB_CREDS\" -X POST --data \'{ \"tag_name\": \"${TAG_NAME}\", \"name\": \"${TAG_NAME}\", \"body\": \"Insert features here.\", \"draft\": true }\' https://api.github.com/repos/Graylog2/graylog-project-cli/releases"
             def props = readJSON text: RELEASE_DATA
             env.RELEASE_ID = props.id

             sh 'curl -fsSL -H "Authorization: token $GITHUB_CREDS" -H "Content-Type: application/octet-stream" --data-binary @graylog-project.linux https://uploads.github.com/repos/Graylog2/graylog-project-cli/releases/$RELEASE_ID/assets?name=graylog-project.linux'
             sh 'curl -fsSL -H "Authorization: token $GITHUB_CREDS" -H "Content-Type: application/octet-stream" --data-binary @graylog-project.darwin-amd64 https://uploads.github.com/repos/Graylog2/graylog-project-cli/releases/$RELEASE_ID/assets?name=graylog-project.darwin-amd64'
             sh 'curl -fsSL -H "Authorization: token $GITHUB_CREDS" -H "Content-Type: application/octet-stream" --data-binary @graylog-project.darwin-arm64 https://uploads.github.com/repos/Graylog2/graylog-project-cli/releases/$RELEASE_ID/assets?name=graylog-project.darwin-arm64'
             sh 'curl -fsSL -H "Authorization: token $GITHUB_CREDS" -H "Content-Type: application/octet-stream" --data-binary @graylog-project.windows-amd64.exe https://uploads.github.com/repos/Graylog2/graylog-project-cli/releases/$RELEASE_ID/assets?name=graylog-project.windows-amd64.exe'
           }
         }
      }
   }

   post
   {
     success
     {
        archiveArtifacts '*.linux'
        archiveArtifacts '*.darwin-amd64'
        archiveArtifacts '*.darwin-arm64'
        archiveArtifacts '*.exe'
     }

     cleanup
     {
       cleanWs()
     }
   }
}
