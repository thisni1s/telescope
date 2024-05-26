@load /cfg/scan.zeek

redef   Analyzer::disable_all = T;
redef   Analyzer::Logging::enable = F;
redef   LogAscii::output_to_stdout = F;

event zeek_init() {
    
    # Disable all other log streams except for Conn
    Log::disable_stream(HTTP::LOG);
    Log::disable_stream(DNS::LOG);
    Log::disable_stream(FTP::LOG);
    Log::disable_stream(SSL::LOG);
    Log::disable_stream(SMTP::LOG);
    Log::disable_stream(SSH::LOG);
    Log::disable_stream(PE::LOG);
    Log::disable_stream(DHCP::LOG);
    Log::disable_stream(NTP::LOG);
    Log::disable_stream(IRC::LOG);
    Log::disable_stream(RDP::LOG);
    Log::disable_stream(QUIC::LOG);
    Log::disable_stream(DPD::LOG);
    Log::disable_stream(OCSP::LOG);
    Log::disable_stream(RADIUS::LOG);
    Log::disable_stream(SIP::LOG);
    Log::disable_stream(SNMP::LOG);
    Log::disable_stream(X509::LOG);
    Log::disable_stream(Files::LOG);
    Log::disable_stream(PacketFilter::LOG);
    Log::disable_stream(Reporter::LOG);
    Log::disable_stream(Tunnel::LOG);
    Log::disable_stream(Weird::LOG);
    Log::disable_stream(KRB::LOG);
    Log::disable_stream(Conn::LOG);
    # You can disable other logs as needed.

    # Enable the conn log stream
    #Log::enable_stream(Conn::LOG);

    local f = Log::get_filter(Notice::LOG, "default");
    # Use tab-separated-value mode
    f$config = table(["tsv"] = "T");
    Log::add_filter(Notice::LOG, f);


    #local g = Log::get_filter(Conn::LOG, "default");
    # Use tab-separated-value mode
    #g$config = table(["tsv"] = "T");
    #Log::add_filter(Conn::LOG, g);
}



redef Notice::ignored_types += {Site::New_Used_Address_Space};
redef Scan::scan_threshold=5;

