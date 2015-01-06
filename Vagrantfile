# -*- mode: ruby -*-
# vi: set ft=ruby :

ENV['VAGRANT_DEFAULT_PROVIDER'] = 'vmware_fusion'

require File.dirname(__FILE__) + '/lib/vagrant-common.rb'

pxc_version = "56"

# Node group counts and aws security groups (if using aws provider)
pxc_nodes = 2
pxc_node_name_prefix = "node"

cluster_address = 'gcomm://' + Array.new( pxc_nodes ){ |i| pxc_node_name_prefix + (i+1).to_s }.join(',')


Vagrant.configure("2") do |config|
	config.vm.box = "perconajayj/centos-x86_64"
	config.ssh.username = "root"

  # Create the PXC nodes
  (1..pxc_nodes).each do |i|
    name = pxc_node_name_prefix + i.to_s
    config.vm.define name do |node_config|
      node_config.vm.hostname = name
      node_config.vm.network :private_network, type: "dhcp"
      node_config.vm.provision :hostmanager
      
      # Provisioners
      provision_puppet( node_config, "pxc_server.pp" ) { |puppet| 
        puppet.facter = {
          # PXC setup
          "percona_server_version"  => pxc_version,
          'innodb_buffer_pool_size' => '128M',
          'innodb_log_file_size' => '64M',
          'innodb_flush_log_at_trx_commit' => '0',
          'pxc_bootstrap_node' => (i == 1 ? true : false ),
          'wsrep_cluster_address' => cluster_address,
          'wsrep_provider_options' => 'gcache.size=128M; gcs.fc_limit=128',
          
          # Sysbench setup
          'sysbench_load' => (i == 1 ? true : false ),
          'tables' => 1,
          'rows' => 100000,
          'threads' => 8,
          'tx_rate' => 10,
          
          # PCT setup
          'percona_agent_api_key' => ENV['PERCONA_AGENT_API_KEY']
        }
      }

      provider_vmware( name, node_config, 256 ) { |vb, override|
        provision_puppet( override, "pxc_server.pp" ) {|puppet|
          puppet.facter = {
            'default_interface' => 'eth1',
            
            # PXC Setup
            'datadir_dev' => 'dm-2',
          }
        }
      }
    end
  end
end