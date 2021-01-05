for dir in `ls`;do
  if [ -f ${dir}/pom.xml ] ;then
  bash -c "cd $dir;mvn package"
  fi
done