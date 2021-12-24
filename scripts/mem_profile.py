import os
import argparse
import json
import subprocess
import shlex
import shutil

def merge_results(outfolder):
  all_results = []
  for f in os.listdir(outfolder):
    result_file = os.path.join(outfolder, f)
    if result_file.endswith('_mem.json') and 'all_results' not in result_file:
      with open(result_file, 'r') as res:
        j = json.load(res)
        all_results.append(j)
  
  with open(os.path.join(outfolder, 'all_results_mem.json'), 'w') as outfile:
    json.dump(all_results, outfile)


def profile_pcap(fname, outfile):
  mem_profile = {
    'nmconfig_general': [],
    'nmconfig_tcp': [],
    'nmconfig_video': []
  }
  for config_file in ['../test/config/trconfig_general.json', '../test/config/trconfig_tcp.json', '../test/config/trconfig_video.json']:
    os.mkdir('/tmp/trtmp')
    cmd = "go run mem_profile.go -trace {} -folder {} -conf {}".format(fname, '/tmp/trtmp/', config_file)
    print("Running profile: {}".format(cmd))
    args = shlex.split(cmd)
    output = subprocess.check_output(args)
    mem_values = []
    for f in os.listdir('/tmp/trtmp'):
      args = shlex.split("go tool pprof -inuse_space -text -unit b --nodefraction=0 {}".format(os.path.join('/tmp/trtmp/', f)))
      mem = subprocess.check_output(args)
      tot = 0.0
      for line in mem.split('\n'):
        if '(*FlowCache).addPacket' in line:
          # print line
          r = line.split()[3]
          mb = float(r.rstrip('B'))
          tot += mb
          # print r, mb
      mem_values.append(tot)
    for key in mem_profile.keys():
      if key in config_file:
        mem_profile[key] = mem_values
        break
    shutil.rmtree('/tmp/trtmp')
  with open(outfile, 'w') as f:
    json.dump(mem_profile, f)
    f.close()


def run_profiling(folder, outfolder):
  for f in os.listdir(folder):
    pcap_file = os.path.join(folder, f)
    if pcap_file.endswith('.pcap'):
      if not os.path.exists(os.path.join(outfolder, f+'_mem.json')):
        
        profile_pcap(pcap_file, os.path.join(outfolder, f+'_mem.json'))
        
    
    
def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('-f', '--folder', type=str, default="data", help="Folder where data is stored")
  parser.add_argument('-o', '--outfolder', type=str, default="results", help="Folder where data is stored")
  args = vars(parser.parse_args())
  
  run_profiling(args['folder'], args['outfolder'])
  merge_results(args['outfolder'])
  


if __name__=='__main__':
  main()