pipeline 
{
   agent { label 'linux' }
   
   tools
   {
     go 'Go'
   }
   
   stages 
   {
      stage('Build') 
      {
         steps {
            git 'https://github.com/Graylog2/graylog-project-cli.git'

            sh "make"
         }

         post 
         {
            success 
            {
               archiveArtifacts '*.linux'
               archiveArtifacts '*.darwin'
            }
         }
      }
   }
}

