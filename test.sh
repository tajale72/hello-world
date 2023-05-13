myfunction() {
    # Define variable with current date/time
    local basename="$(pwd)"
    local bas_file="$(basename ${basename})"
    local go_test_coverage="/tmp/go_test_coverage/romit.out"
    
    go test -cover | grep coverage > "${go_test_coverage}"
    # Output variable
    echo $basename
    echo $bas_file

}
