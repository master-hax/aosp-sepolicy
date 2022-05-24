#include <getopt.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>

#include <cil/cil.h>
#include <cil/android.h>
#include <sepol/policydb.h>

void usage(const char *prog)
{
    printf("Usage: %s [OPTION]... FILE...\n", prog);
    printf("\n");
    printf("Options:\n");
    printf("  -b, --base=<file>          (required) base binary policy.\n");
    printf("  -o, --output=<file>        (required) write binary policy to <file>\n");
    printf("  -v, --verbose              increment verbosity level\n");
    printf("  -h, --help                 display usage information\n");
    exit(1);
}

/*
 * read_cil_files - Initialize db and parse CIL input files.
 */
static int read_cil_files(struct cil_db **db, char **paths,
                          unsigned int n_files)
{
    int rc = SEPOL_ERR;
    FILE *file;
    struct stat filedata;
    uint32_t file_size;
    char *buff = NULL;

    for (int i = 0; i < n_files; i++) {
        char *path = paths[i];

        file = fopen(path, "r");
        if (!file) {
            rc = SEPOL_ERR;
            fprintf(stderr, "Could not open file: %s.\n", path);
            goto file_err;
        }

        rc = stat(path, &filedata);
        if (rc == -1) {
            fprintf(stderr, "Could not stat file: %s - %s.\n", path, strerror(errno));
            goto err;
        }

        file_size = filedata.st_size;
        buff = malloc(file_size);
        if (buff == NULL) {
            fprintf(stderr, "OOM!\n");
            rc = SEPOL_ERR;
            goto err;
        }

        rc = fread(buff, file_size, 1, file);
        if (rc != 1) {
            fprintf(stderr, "Failure reading file: %s.\n", path);
            rc = SEPOL_ERR;
            goto err;
        }
        fclose(file);
        file = NULL;

        /* create parse_tree */
        rc = cil_add_file(*db, path, buff, file_size);
        if (rc != SEPOL_OK) {
            fprintf(stderr, "Failure adding %s to parse tree.\n", path);
            goto parse_err;
        }
        free(buff);
    }

    return SEPOL_OK;
err:
    fclose(file);
parse_err:
    free(buff);
file_err:
    cil_db_destroy(db);
    return rc;
}

int main(int argc, char *argv[])
{
    int rc = SEPOL_ERR;
    sepol_policydb_t *pdb = NULL;
    struct sepol_policy_file *pf_base = NULL;
    FILE *binary_base = NULL;
    char *base = NULL;
    struct cil_db *incremental_db = NULL;
    struct sepol_policy_file *pf_out = NULL;
    FILE *binary_out = NULL;
    char *output = NULL;
    struct stat binarydata;
    uint32_t binary_size;
    int opt_char;
    int opt_index = 0;
    enum cil_log_level log_level = CIL_ERR;
    static struct option long_opts[] = {{"base", required_argument, 0, 'b'},
                                        {"output", required_argument, 0, 'o'},
                                        {"verbose", no_argument, 0, 'v'},
                                        {"help", no_argument, 0, 'h'},
                                        {0, 0, 0, 0}};

    while (1) {
        opt_char = getopt_long(argc, argv, "b:o:vh", long_opts, &opt_index);
		if (opt_char == -1) {
			break;
		}
        switch (opt_char) {
        case 'b':
            base = strdup(optarg);
            break;
        case 'o':
            free(output);
            output = strdup(optarg);
            break;
        case 'v':
            log_level++;
            break;
        case 'h':
            usage(argv[0]);
        default:
            fprintf(stderr, "Unsupported option: %s.\n", optarg);
            usage(argv[0]);
        }
    }
    if (base == NULL || output == NULL) {
        fprintf(stderr, "Please specify required arguments.\n");
        usage(argv[0]);
    }

    cil_set_log_level(log_level);

    /*
     * Read the base binary policy.
     */
    binary_base = fopen(base, "r");
    if (!binary_base) {
        fprintf(stderr, "Could not open base binary file: %s.\n", base);
        rc = SEPOL_ERR;
        goto exit;
    }

    rc = stat(base, &binarydata);
    if (rc == -1) {
        fprintf(stderr, "Could not stat base binary file: %s.\n", base);
        rc = SEPOL_ERR;
        goto exit;
    }
    binary_size = binarydata.st_size;
    if (!binary_size) {
        fprintf(stderr, "No binary size.\n");
        binary_base = NULL;
        rc = SEPOL_ERR;
        goto exit;
    }

    rc = sepol_policy_file_create(&pf_base);
    if (rc != 0) {
        fprintf(stderr, "Failed to create policy file: %d.\n", rc);
        goto exit;
    }
    sepol_policy_file_set_fp(pf_base, binary_base);

    rc = sepol_policydb_create(&pdb);
    if (rc != 0)
    {
        fprintf(stderr, "Could not create policy db: %d.\n", rc);
        goto exit;
    }

    rc = sepol_policydb_read(pdb, pf_base);
    if (rc != 0) {
        fprintf(stderr, "Failed to read binary policy: %d.\n", rc);
        goto exit;
    }

    /*
     * Initialize db and amend the policyd db.
     */
    cil_db_init(&incremental_db);
    rc = read_cil_files(&incremental_db, argv + optind, argc - optind);
    if (rc != SEPOL_OK) {
        fprintf(stderr, "Failed to read CIL files: %d.\n", rc);
        goto exit;
    }

    rc = cil_compile(incremental_db);
    if (rc != SEPOL_OK) {
        fprintf(stderr, "Failed to compile cildb: %d.\n", rc);
        goto exit;
    }

    rc = cil_amend_policydb(incremental_db, pdb);
    if (rc != SEPOL_OK) {
        fprintf(stderr, "Failed to build policydb.\n");
        goto exit;
    }

    /*
     * Write the result to file.
     */
    binary_out = fopen(output, "w");
    if (binary_out == NULL) {
        fprintf(stderr, "Failure opening binary %s file for writing.\n", output);
        rc = SEPOL_ERR;
        goto exit;
    }

    rc = sepol_policy_file_create(&pf_out);
    if (rc != 0) {
        fprintf(stderr, "Failed to create policy file: %d.\n", rc);
        goto exit;
    }
    sepol_policy_file_set_fp(pf_out, binary_out);

    rc = sepol_policydb_write(pdb, pf_out);
    if (rc != 0) {
        fprintf(stderr, "failed to write binary policy: %d.\n", rc);
        goto exit;
    }

exit:
    if (binary_base != NULL) {
        fclose(binary_base);
    }
    if (binary_out != NULL) {
        fclose(binary_out);
    }
    sepol_policydb_free(pdb);
    cil_db_destroy(&incremental_db);
    sepol_policy_file_free(pf_base);
    sepol_policy_file_free(pf_out);
    free(output);
    return rc;
}
