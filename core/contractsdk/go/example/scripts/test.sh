for contract in `ls tests`;do
  xdev test tests/$contract || exit -1
done