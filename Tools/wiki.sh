cd ..
rm -rf wiki
mkdir wiki
for file in `find . -name "*.go"`
do
    echo $file
    python3 Tools/nyawiki.py $file >$file.md
done
for file in `find . -name "*.go.md"`
do
    mv $file ./wiki
done
cd Tools