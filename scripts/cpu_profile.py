import os
import argparse
import json
import subprocess
import shlex

def merge_results(outfolder):
  all_results = []
  for f in os.listdir(outfolder):
    result_file = os.path.join(outfolder, f)
    if result_file.endswith('_cpu.json') and 'all_results' not in result_file:
      with open(result_file, 'r') as res:
        j = json.load(res)
        all_results.append(j)
  
  with open(os.path.join(outfolder, 'all_results_cpu.json'), 'w') as outfile:
    json.dump(all_results, outfile)


def profile_pcap(fname, outfile):
  cmd = "go run cpu_profile.go -trace {}".format(fname)
  print("Running profile: {}".format(cmd))
  args = shlex.split(cmd)
  output = subprocess.check_output(args)
  with open(outfile, 'w') as f:
    f.write(output)
    f.close()


def run_profiling(folder, outfolder):
  for f in os.listdir(folder):
    pcap_file = os.path.join(folder, f)
    if pcap_file.endswith('.pcap'):
      if not os.path.exists(os.path.join(outfolder, f+'_cpu.json')):
        profile_pcap(pcap_file, os.path.join(outfolder, f+'_cpu.json'))
    
    
def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('-f', '--folder', type=str, default="data", help="Folder where data is stored")
  parser.add_argument('-o', '--outfolder', type=str, default="results", help="Folder where data is stored")
  args = vars(parser.parse_args())
  
  run_profiling(args['folder'], args['outfolder'])
  merge_results(args['outfolder'])
  


if __name__=='__main__':
  main()