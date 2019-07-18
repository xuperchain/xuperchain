#include <inttypes.h>
#include "xchain/xchain.h"
#include "table/types.h"
#include "table/table.cc"
#include "pb/anchor.pb.h"

//data anchoring
struct Anchor : public xchain::Contract {
public:
    Anchor(): _entity(this->context(), "entity") {}

    struct entity: public anchor::Entity {
        DEFINE_ROWKEY(name());
        DEFINE_INDEX_BEGIN(2)
            DEFINE_INDEX_ADD(0, id(), name())
            DEFINE_INDEX_ADD(1, name(), desc())
        DEFINE_INDEX_END();
    };
private:
    xchain::cdt::Table<entity> _entity;

public:
    decltype(_entity)& get_entity() {
        return _entity;
    }
};

//初始化
DEFINE_METHOD(Anchor, initialize) {
    xchain::Context* ctx = self.context();
    const std::string& id= ctx->arg("id");
    const std::string& name = ctx->arg("name");
    const std::string& desc = ctx->arg("desc");

    Anchor::entity ent;
    ent.set_id(std::stoll(id));
    ent.set_name(name.c_str());
    ent.set_desc(desc);
    self.get_entity().put(ent);
    ctx->ok("initialize succeed");
}

DEFINE_METHOD(Anchor, get) {
    xchain::Context* ctx = self.context();
    const std::string& name = ctx->arg("key");
    Anchor::entity ent;
    if (self.get_entity().find(name.c_str(), &ent)) {
        ctx->ok(ent.to_str());
        return;
    }
    ctx->error("can not find " + name);
}

DEFINE_METHOD(Anchor, set) {
    xchain::Context* ctx = self.context();
    const std::string& id= ctx->arg("id");
    const std::string& name = ctx->arg("name");
    const std::string& desc = ctx->arg("desc");

    Anchor::entity ent;
    ent.set_id(std::stoll(id));
    ent.set_name(name.c_str());
    ent.set_desc(desc);
    self.get_entity().put(ent);
    ctx->ok("done");
}

DEFINE_METHOD(Anchor, scan) {
    xchain::Context* ctx = self.context();
    const std::string& id= ctx->arg("index");

    auto it = self.get_entity().scan(id);
    Anchor::entity ent;
    int i = 0;
    while(!it.end()) {
        if (it.get(&ent)) {
            std::cout << "id: " << ent.id()<< std::endl;
            std::cout << "name: " << ent.name()<< std::endl;
            std::cout << "desc: " << ent.desc()<< std::endl;
            i += 1;
        }
        if (!it.next()) {
            std::cout << "error !!!" << std::endl;
            break;
        }
    }
    ctx->ok(std::to_string(i));
}
