#ifndef RINER_RUNTIME_H
#define RINER_RUNTIME_H

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

/* ── Basic types ── */
typedef int64_t  riner_int;
typedef double   riner_float;
typedef char*    riner_string;
typedef int      riner_bool;
typedef void*    riner_any;

#define RINER_TRUE  1
#define RINER_FALSE 0
#define RINER_NIL   NULL

/* ── String ── */
static riner_string riner_str_concat(riner_string a, riner_string b) {
    size_t la = strlen(a), lb = strlen(b);
    riner_string out = (riner_string)malloc(la + lb + 1);
    memcpy(out, a, la);
    memcpy(out + la, b, lb + 1);
    return out;
}

static riner_string riner_int_to_str(riner_int n) {
    char buf[32];
    snprintf(buf, sizeof(buf), "%lld", (long long)n);
    return strdup(buf);
}

static riner_string riner_float_to_str(riner_float f) {
    char buf[64];
    snprintf(buf, sizeof(buf), "%g", f);
    return strdup(buf);
}

static riner_string riner_bool_to_str(riner_bool b) {
    return b ? strdup("true") : strdup("false");
}

/* ── Print ── */
static void riner_print(riner_string s) {
    printf("%s\n", s);
}

/* ── Array ── */
typedef struct {
    riner_any* data;
    size_t     len;
    size_t     cap;
} riner_array;

static riner_array* riner_array_new() {
    riner_array* a = (riner_array*)malloc(sizeof(riner_array));
    a->data = NULL;
    a->len  = 0;
    a->cap  = 0;
    return a;
}

static void riner_array_push(riner_array* a, riner_any val) {
    if (a->len >= a->cap) {
        a->cap = a->cap == 0 ? 4 : a->cap * 2;
        a->data = (riner_any*)realloc(a->data, a->cap * sizeof(riner_any));
    }
    a->data[a->len++] = val;
}

static riner_any riner_array_get(riner_array* a, riner_int i) {
    if (i < 0 || (size_t)i >= a->len) {
        fprintf(stderr, "runtime error: index out of bounds\n");
        exit(1);
    }
    return a->data[i];
}

static void riner_array_set(riner_array* a, riner_int i, riner_any val) {
    if (i < 0 || (size_t)i >= a->len) {
        fprintf(stderr, "runtime error: index out of bounds\n");
        exit(1);
    }
    a->data[i] = val;
}

static riner_int riner_array_len(riner_array* a) {
    return (riner_int)a->len;
}

/* ── Map ── */
typedef struct riner_map_entry {
    riner_string key;
    riner_any    val;
    struct riner_map_entry* next;
} riner_map_entry;

typedef struct {
    riner_map_entry** buckets;
    size_t            size;
} riner_map;

static riner_map* riner_map_new() {
    riner_map* m = (riner_map*)malloc(sizeof(riner_map));
    m->size = 16;
    m->buckets = (riner_map_entry**)calloc(m->size, sizeof(riner_map_entry*));
    return m;
}

static size_t riner_map_hash(riner_string key, size_t size) {
    size_t h = 5381;
    while (*key) h = h * 33 + (unsigned char)*key++;
    return h % size;
}

static void riner_map_set(riner_map* m, riner_string key, riner_any val) {
    size_t h = riner_map_hash(key, m->size);
    riner_map_entry* e = m->buckets[h];
    while (e) {
        if (strcmp(e->key, key) == 0) { e->val = val; return; }
        e = e->next;
    }
    riner_map_entry* ne = (riner_map_entry*)malloc(sizeof(riner_map_entry));
    ne->key  = strdup(key);
    ne->val  = val;
    ne->next = m->buckets[h];
    m->buckets[h] = ne;
}

static riner_any riner_map_get(riner_map* m, riner_string key) {
    size_t h = riner_map_hash(key, m->size);
    riner_map_entry* e = m->buckets[h];
    while (e) {
        if (strcmp(e->key, key) == 0) return e->val;
        e = e->next;
    }
    return RINER_NIL;
}

/* ── Panic ── */
static void riner_panic(riner_string msg) {
    fprintf(stderr, "runtime error: %s\n", msg);
    exit(1);
}

#endif /* RINER_RUNTIME_H */